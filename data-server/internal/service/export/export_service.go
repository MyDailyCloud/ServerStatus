package export

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// ExportService 数据导出服务
type ExportService struct {
	serverRepo  repository.ServerRepository
	historyRepo repository.HistoryRepository
	cacheRepo   repository.CacheRepository
	logger      logger.Logger
}

// NewExportService 创建数据导出服务
func NewExportService(
	serverRepo repository.ServerRepository,
	historyRepo repository.HistoryRepository,
	cacheRepo repository.CacheRepository,
	logger logger.Logger,
) *ExportService {
	return &ExportService{
		serverRepo:  serverRepo,
		historyRepo: historyRepo,
		cacheRepo:   cacheRepo,
		logger:      logger,
	}
}

// ExportFormat 导出格式
type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatJSON ExportFormat = "json"
)

// ExportRequest 导出请求
type ExportRequest struct {
	ProjectKey   string           `json:"project_key"`
	Hostnames    []string         `json:"hostnames,omitempty"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	Format       ExportFormat     `json:"format"`
	IncludeTypes []ExportDataType `json:"include_types"`
	Limit        int              `json:"limit,omitempty"`
	Offset       int              `json:"offset,omitempty"`
}

// ExportDataType 导出数据类型
type ExportDataType string

const (
	DataTypeServerInfo ExportDataType = "server_info"
	DataTypeSystemInfo ExportDataType = "system_info"
	DataTypeHistory    ExportDataType = "history"
)

// ExportResult 导出结果
type ExportResult struct {
	Filename    string                 `json:"filename"`
	ContentType string                 `json:"content_type"`
	Size        int64                  `json:"size"`
	RecordCount int                    `json:"record_count"`
	ExportTime  time.Duration          `json:"export_time"`
	Data        io.ReadCloser          `json:"-"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ExportServers 导出服务器信息
func (s *ExportService) ExportServers(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	startTime := time.Now()

	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid export request: %w", err)
	}

	// 获取服务器数据
	servers, err := s.getServerData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get server data: %w", err)
	}

	// 根据格式导出数据
	var result *ExportResult
	switch req.Format {
	case FormatCSV:
		result, err = s.exportServersToCSV(ctx, servers, req)
	case FormatJSON:
		result, err = s.exportServersToJSON(ctx, servers, req)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to export data: %w", err)
	}

	result.ExportTime = time.Since(startTime)
	result.Metadata = map[string]interface{}{
		"export_type":   "servers",
		"project_key":   req.ProjectKey,
		"server_count":  len(servers),
		"start_time":    req.StartTime,
		"end_time":      req.EndTime,
		"include_types": req.IncludeTypes,
	}

	s.logger.WithFields(map[string]interface{}{
		"format":       req.Format,
		"server_count": len(servers),
		"duration":     result.ExportTime,
		"size":         result.Size,
	}).Info("Server data exported successfully")

	return result, nil
}

// ExportHistory 导出历史数据
func (s *ExportService) ExportHistory(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	startTime := time.Now()

	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid export request: %w", err)
	}

	// 获取历史数据
	historyData, err := s.getHistoryData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get history data: %w", err)
	}

	// 根据格式导出数据
	var result *ExportResult
	switch req.Format {
	case FormatCSV:
		result, err = s.exportHistoryToCSV(ctx, historyData, req)
	case FormatJSON:
		result, err = s.exportHistoryToJSON(ctx, historyData, req)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to export history data: %w", err)
	}

	result.ExportTime = time.Since(startTime)
	result.Metadata = map[string]interface{}{
		"export_type":  "history",
		"project_key":  req.ProjectKey,
		"record_count": len(historyData),
		"start_time":   req.StartTime,
		"end_time":     req.EndTime,
	}

	s.logger.WithFields(map[string]interface{}{
		"format":       req.Format,
		"record_count": len(historyData),
		"duration":     result.ExportTime,
		"size":         result.Size,
	}).Info("History data exported successfully")

	return result, nil
}

// validateRequest 验证导出请求
func (s *ExportService) validateRequest(req *ExportRequest) error {
	if req.ProjectKey == "" {
		return fmt.Errorf("project_key is required")
	}

	if req.StartTime.After(req.EndTime) {
		return fmt.Errorf("start_time cannot be after end_time")
	}

	if req.StartTime.AddDate(0, 0, 365).Before(req.EndTime) {
		return fmt.Errorf("export time range cannot exceed 365 days")
	}

	if req.Format == "" {
		req.Format = FormatCSV // 默认CSV格式
	}

	if req.Limit <= 0 || req.Limit > 100000 {
		req.Limit = 10000 // 默认限制1万条记录
	}

	if len(req.IncludeTypes) == 0 {
		req.IncludeTypes = []ExportDataType{DataTypeServerInfo, DataTypeSystemInfo}
	}

	return nil
}

// getServerData 获取服务器数据
func (s *ExportService) getServerData(ctx context.Context, req *ExportRequest) ([]*models.ServerInfo, error) {
	var servers []*models.ServerInfo

	if len(req.Hostnames) > 0 {
		// 根据主机名查询多个服务器的最新数据
		for _, hostname := range req.Hostnames {
			// 获取指定时间范围内的历史数据
			historyData, err := s.historyRepo.GetHistoryByTimeRange(ctx, hostname, req.ProjectKey, req.StartTime, req.EndTime)
			if err != nil {
				s.logger.WithError(err).WithField("hostname", hostname).Warn("Failed to get history for hostname")
				continue
			}
			if len(historyData) == 0 {
				continue
			}
			latest := s.convertHistoryDataToSystemInfo(historyData[len(historyData)-1])
			servers = append(servers, &models.ServerInfo{
				SessionID:  latest.SessionID,
				Hostname:   latest.Hostname,
				ProjectKey: latest.ProjectKey,
				Latest:     latest,
			})
		}
	} else {
		if s.serverRepo == nil {
			return nil, fmt.Errorf("server repository is not configured")
		}
		serversFromRepo, err := s.serverRepo.GetAllServers(ctx, req.ProjectKey, req.Offset, req.Limit)
		if err != nil {
			return nil, err
		}
		for _, server := range serversFromRepo {
			// 如果Latest为空，尝试从History中取最新一条
			if server.Latest == nil && len(server.History) > 0 {
				server.Latest = server.History[len(server.History)-1]
			}
			// 如果依然为空，使用基本信息填充
			if server.Latest == nil {
				server.Latest = &models.SystemInfo{
					Hostname:   server.Hostname,
					SessionID:  server.SessionID,
					ProjectKey: server.ProjectKey,
					Timestamp:  server.UpdatedAt,
					CPU: models.CPUInfo{
						CoreCount: server.CPUCores,
					},
					Memory: models.MemInfo{
						Total: server.MemoryTotal,
					},
					Disk: models.DiskInfo{
						Total: server.DiskTotal,
					},
					Network: models.NetInfo{},
				}
			}
			servers = append(servers, server)
		}
	}

	// 按时间排序（基于最新数据）
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Latest.Timestamp.Before(servers[j].Latest.Timestamp)
	})

	// 应用分页
	if req.Offset >= len(servers) {
		return []*models.ServerInfo{}, nil
	}

	end := req.Offset + req.Limit
	if end > len(servers) {
		end = len(servers)
	}

	return servers[req.Offset:end], nil
}

// getHistoryData 获取历史数据
func (s *ExportService) getHistoryData(ctx context.Context, req *ExportRequest) ([]*models.SystemInfo, error) {
	var allHistory []*models.SystemInfo

	// 如果指定了主机名，批量获取历史数据
	if len(req.Hostnames) > 0 {
		for _, hostname := range req.Hostnames {
			historyData, err := s.historyRepo.GetHistoryByTimeRange(ctx, hostname, req.ProjectKey, req.StartTime, req.EndTime)
			if err != nil {
				s.logger.WithError(err).WithField("hostname", hostname).Warn("Failed to get history for hostname")
				continue
			}
			// 转换HistoryData为SystemInfo
			for _, data := range historyData {
				systemInfo := s.convertHistoryDataToSystemInfo(data)
				allHistory = append(allHistory, systemInfo)
			}
		}
	} else {
		if s.serverRepo == nil {
			return nil, fmt.Errorf("server repository is not configured")
		}
		// 先获取项目下的服务器列表，然后分别拉取历史数据
		serverLimit := req.Limit
		if serverLimit > 1000 {
			serverLimit = 1000
		}
		servers, err := s.serverRepo.GetAllServers(ctx, req.ProjectKey, req.Offset, serverLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to list servers: %w", err)
		}
		for _, server := range servers {
			historyData, err := s.historyRepo.GetHistoryByTimeRange(ctx, server.SessionID, req.ProjectKey, req.StartTime, req.EndTime)
			if err != nil {
				s.logger.WithError(err).WithField("server", server.SessionID).Warn("Failed to get history for server")
				continue
			}
			for _, data := range historyData {
				allHistory = append(allHistory, s.convertHistoryDataToSystemInfo(data))
			}
		}
	}

	// 按时间排序
	sort.Slice(allHistory, func(i, j int) bool {
		return allHistory[i].Timestamp.Before(allHistory[j].Timestamp)
	})

	// 应用分页
	if req.Offset >= len(allHistory) {
		return []*models.SystemInfo{}, nil
	}

	end := req.Offset + req.Limit
	if end > len(allHistory) {
		end = len(allHistory)
	}

	return allHistory[req.Offset:end], nil
}

// convertHistoryDataToSystemInfo 将HistoryData转换为SystemInfo
func (s *ExportService) convertHistoryDataToSystemInfo(data *models.HistoryData) *models.SystemInfo {
	// 计算内存总量（如果只有已用和使用率）
	memoryTotal := data.MemoryUsed
	if data.MemoryUsage > 0 && data.MemoryUsage <= 100 {
		memoryTotal = uint64(float64(data.MemoryUsed) / (data.MemoryUsage / 100.0))
	}

	// 计算磁盘总量（如果只有已用和使用率）
	diskTotal := data.DiskUsed
	if data.DiskUsage > 0 && data.DiskUsage <= 100 {
		diskTotal = uint64(float64(data.DiskUsed) / (data.DiskUsage / 100.0))
	}

	return &models.SystemInfo{
		Hostname:   data.Hostname,
		SessionID:  data.SessionID,
		Timestamp:  data.Timestamp,
		ProjectKey: data.ProjectKey,
		CPU: models.CPUInfo{
			UsagePercent: data.CPUUsage,
		},
		Memory: models.MemInfo{
			Total:        memoryTotal,
			Used:         data.MemoryUsed,
			Free:         memoryTotal - data.MemoryUsed,
			UsagePercent: data.MemoryUsage,
		},
		Disk: models.DiskInfo{
			Total:        diskTotal,
			Used:         data.DiskUsed,
			Free:         diskTotal - data.DiskUsed,
			UsagePercent: data.DiskUsage,
		},
		Network: models.NetInfo{
			BytesRecv: data.NetworkRx,
			BytesSent: data.NetworkTx,
		},
		// 注意：HistoryData中有GPU信息，但SystemInfo中的GPUInfo结构不同
		// 这里需要根据实际的GPUInfo结构调整
	}
}

// exportServersToCSV 导出服务器信息为CSV
func (s *ExportService) exportServersToCSV(ctx context.Context, servers []*models.ServerInfo, req *ExportRequest) (*ExportResult, error) {
	var records [][]string

	// 添加表头
	headers := []string{
		"SessionID", "Hostname", "ProjectKey", "OS", "Architecture",
		"CPUCores", "MemoryTotal", "DiskTotal", "Uptime", "BootTime",
		"CreatedAt", "UpdatedAt", "CPUUsage", "MemoryUsed", "MemoryAvailable",
		"DiskUsed", "DiskAvailable", "NetworkRx", "NetworkTx", "LoadAvg",
		"ProcessCount", "LastUpdate",
	}

	records = append(records, headers)

	// 添加数据行
	for _, server := range servers {
		latest := server.Latest
		if latest == nil {
			latest = &models.SystemInfo{
				Hostname:   server.Hostname,
				SessionID:  server.SessionID,
				ProjectKey: server.ProjectKey,
				Timestamp:  server.UpdatedAt,
			}
		}
		record := []string{
			server.SessionID,
			server.Hostname,
			server.ProjectKey,
			server.OS,
			server.Arch,
			fmt.Sprintf("%d", server.CPUCores),
			s.formatBytes(server.MemoryTotal),
			s.formatBytes(server.DiskTotal),
			fmt.Sprintf("%d", server.Uptime),
			server.BootTime.Format(time.RFC3339),
			server.CreatedAt.Format(time.RFC3339),
			server.UpdatedAt.Format(time.RFC3339),
			fmt.Sprintf("%.2f", latest.CPU.UsagePercent),
			s.formatBytes(latest.Memory.Used),
			s.formatBytes(latest.Memory.Free),
			s.formatBytes(latest.Disk.Used),
			s.formatBytes(latest.Disk.Free),
			s.formatBytes(latest.Network.BytesRecv),
			s.formatBytes(latest.Network.BytesSent),
			"", // LoadAvg 暂无数据源
			"", // ProcessCount 暂无数据源
			latest.Timestamp.Format(time.RFC3339),
		}

		records = append(records, record)
	}

	// 生成CSV内容
	csvContent, err := s.generateCSVContent(records)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("servers_export_%s_%d.csv", req.ProjectKey, time.Now().Unix())

	return &ExportResult{
		Filename:    filename,
		ContentType: "text/csv",
		Size:        int64(len(csvContent)),
		RecordCount: len(records) - 1, // 减去表头
		Data:        &StringReadCloser{Data: csvContent},
	}, nil
}

// exportHistoryToCSV 导出历史数据为CSV
func (s *ExportService) exportHistoryToCSV(ctx context.Context, history []*models.SystemInfo, req *ExportRequest) (*ExportResult, error) {
	var records [][]string

	// 添加表头
	headers := []string{
		"Hostname", "SessionID", "ProjectKey", "Timestamp", "CPUUsage", "MemoryTotal", "MemoryUsed", "MemoryUsagePercent",
		"DiskTotal", "DiskUsed", "DiskUsagePercent", "NetworkRx", "NetworkTx",
	}
	records = append(records, headers)

	// 添加数据行
	for _, sys := range history {
		record := []string{
			sys.Hostname,
			sys.SessionID,
			sys.ProjectKey,
			sys.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%.2f", sys.CPU.UsagePercent),
			s.formatBytes(sys.Memory.Total),
			s.formatBytes(sys.Memory.Used),
			fmt.Sprintf("%.2f", sys.Memory.UsagePercent),
			s.formatBytes(sys.Disk.Total),
			s.formatBytes(sys.Disk.Used),
			fmt.Sprintf("%.2f", sys.Disk.UsagePercent),
			s.formatBytes(sys.Network.BytesRecv),
			s.formatBytes(sys.Network.BytesSent),
		}
		records = append(records, record)
	}

	// 生成CSV内容
	csvContent, err := s.generateCSVContent(records)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("history_export_%s_%d.csv", req.ProjectKey, time.Now().Unix())

	return &ExportResult{
		Filename:    filename,
		ContentType: "text/csv",
		Size:        int64(len(csvContent)),
		RecordCount: len(records) - 1, // 减去表头
		Data:        &StringReadCloser{Data: csvContent},
	}, nil
}

// exportServersToJSON 导出服务器信息为JSON
func (s *ExportService) exportServersToJSON(ctx context.Context, servers []*models.ServerInfo, req *ExportRequest) (*ExportResult, error) {
	// 构建导出数据结构
	exportData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"export_time":   time.Now().Format(time.RFC3339),
			"project_key":   req.ProjectKey,
			"total_servers": len(servers),
			"include_types": req.IncludeTypes,
			"start_time":    req.StartTime.Format(time.RFC3339),
			"end_time":      req.EndTime.Format(time.RFC3339),
		},
		"servers": servers,
	}

	// 序列化为JSON（紧凑格式便于匹配）
	jsonData, err := json.Marshal(exportData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal server info to JSON: %w", err)
	}

	filename := fmt.Sprintf("servers_export_%s_%d.json", req.ProjectKey, time.Now().Unix())

	return &ExportResult{
		Filename:    filename,
		ContentType: "application/json",
		Size:        int64(len(jsonData)),
		RecordCount: len(servers),
		Data:        &StringReadCloser{Data: string(jsonData)},
	}, nil
}

// exportHistoryToJSON 导出历史数据为JSON
func (s *ExportService) exportHistoryToJSON(ctx context.Context, history []*models.SystemInfo, req *ExportRequest) (*ExportResult, error) {
	// 导出为简洁的数组，便于直接消费
	jsonData, err := json.Marshal(history)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal history to JSON: %w", err)
	}

	filename := fmt.Sprintf("history_export_%s_%d.json", req.ProjectKey, time.Now().Unix())

	return &ExportResult{
		Filename:    filename,
		ContentType: "application/json",
		Size:        int64(len(jsonData)),
		RecordCount: len(history),
		Data:        &StringReadCloser{Data: string(jsonData)},
	}, nil
}

// generateCSVContent 生成CSV内容
func (s *ExportService) generateCSVContent(records [][]string) (string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return buf.String(), nil
}

// formatBytes 格式化字节数
func (s *ExportService) formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// StringReadCloser 字符串读取器，实现io.ReadCloser接口
type StringReadCloser struct {
	Data string
	pos  int
}

func (rc *StringReadCloser) Read(p []byte) (n int, err error) {
	if rc.pos >= len(rc.Data) {
		return 0, io.EOF
	}
	n = copy(p, rc.Data[rc.pos:])
	rc.pos += n
	return n, nil
}

func (rc *StringReadCloser) Close() error {
	rc.pos = 0
	return nil
}

// GetExportFormats 获取支持的导出格式
func (s *ExportService) GetExportFormats() []ExportFormat {
	return []ExportFormat{FormatCSV, FormatJSON}
}

// GetExportDataTypes 获取支持的数据类型
func (s *ExportService) GetExportDataTypes() []ExportDataType {
	return []ExportDataType{DataTypeServerInfo, DataTypeSystemInfo, DataTypeHistory}
}

// EstimateExportSize 估算导出文件大小
func (s *ExportService) EstimateExportSize(ctx context.Context, req *ExportRequest) (int64, error) {
	// 获取少量样本数据来估算大小
	sampleReq := *req
	sampleReq.Limit = 10
	sampleReq.Offset = 0

	var sampleSize int64
	var recordCount int

	servers, err := s.getServerData(ctx, &sampleReq)
	if err != nil {
		return 0, err
	}

	var result *ExportResult
	switch req.Format {
	case FormatCSV:
		result, err = s.exportServersToCSV(ctx, servers, &sampleReq)
	case FormatJSON:
		result, err = s.exportServersToJSON(ctx, servers, &sampleReq)
	}

	if err != nil {
		return 0, err
	}

	sampleSize = result.Size
	recordCount = len(servers)

	if recordCount == 0 {
		return 0, nil
	}

	// 估算总大小
	avgSizePerRecord := float64(sampleSize) / float64(recordCount)
	estimatedSize := int64(avgSizePerRecord * float64(req.Limit))

	return estimatedSize, nil
}

// containsDataType 检查是否包含指定数据类型
func (s *ExportService) containsDataType(types []ExportDataType, targetType ExportDataType) bool {
	for _, t := range types {
		if t == targetType {
			return true
		}
	}
	return false
}
