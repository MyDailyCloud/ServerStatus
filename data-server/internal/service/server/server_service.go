package server

import (
	"context"
	"fmt"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// ServerService 服务器管理服务实现
type ServerService struct {
	serverRepo     repository.ServerRepository
	historyRepo    repository.HistoryRepository
	cacheRepo      repository.CacheRepository
	logger         logger.Logger
	offlineTimeout time.Duration
}

// NewServerService 创建服务器管理服务
func NewServerService(
	serverRepo repository.ServerRepository,
	historyRepo repository.HistoryRepository,
	cacheRepo repository.CacheRepository,
	logger logger.Logger,
) *ServerService {
	return &ServerService{
		serverRepo:     serverRepo,
		historyRepo:    historyRepo,
		cacheRepo:      cacheRepo,
		logger:         logger,
		offlineTimeout: 5 * time.Minute, // 5分钟无响应视为离线
	}
}

// RegisterServer 注册新服务器
func (s *ServerService) RegisterServer(ctx context.Context, info *models.SystemInfo) error {
	// 验证输入数据
	if err := s.validateSystemInfo(info); err != nil {
		return fmt.Errorf("invalid system info: %w", err)
	}

	// 创建服务器信息
	serverInfo := &models.ServerInfo{
		SessionID:   info.SessionID,
		Hostname:    info.Hostname,
		ProjectKey:  info.ProjectKey,
		OS:          info.OS,
		Arch:        info.Arch,
		CPUCores:    info.CPUInfo.Cores,
		MemoryTotal: info.MemInfo.Total,
		DiskTotal:   info.DiskInfo.Total,
		Uptime:      info.Uptime,
		BootTime:    info.BootTime,
		SystemInfo:  info,
		IsOnline:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存到数据库
	if err := s.serverRepo.CreateServer(ctx, serverInfo); err != nil {
		s.logger.WithError(err).WithField("session_id", info.SessionID).Error("Failed to register server")
		return fmt.Errorf("failed to register server: %w", err)
	}

	// 保存初始历史数据
	if err := s.saveHistoryData(ctx, info); err != nil {
		s.logger.WithError(err).WithField("session_id", info.SessionID).Warn("Failed to save initial history data")
	}

	// 缓存服务器信息
	s.cacheServerInfo(ctx, serverInfo)

	s.logger.WithFields(map[string]interface{}{
		"session_id": info.SessionID,
		"hostname":   info.Hostname,
		"os":         info.OS,
	}).Info("Server registered successfully")

	return nil
}

// UpdateServerStatus 更新服务器状态
func (s *ServerService) UpdateServerStatus(ctx context.Context, info *models.SystemInfo) error {
	// 验证输入数据
	if err := s.validateSystemInfo(info); err != nil {
		return fmt.Errorf("invalid system info: %w", err)
	}

	// 获取现有服务器信息
	serverInfo, err := s.serverRepo.GetServer(ctx, info.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}

	// 更新状态信息
	serverInfo.SystemInfo = info
	serverInfo.UpdatedAt = time.Now()
	serverInfo.IsOnline = true

	// 更新数据库
	if err := s.serverRepo.UpdateServer(ctx, serverInfo); err != nil {
		s.logger.WithError(err).WithField("session_id", info.SessionID).Error("Failed to update server status")
		return fmt.Errorf("failed to update server status: %w", err)
	}

	// 保存历史数据
	if err := s.saveHistoryData(ctx, info); err != nil {
		s.logger.WithError(err).WithField("session_id", info.SessionID).Warn("Failed to save history data")
	}

	// 更新缓存
	s.cacheServerInfo(ctx, serverInfo)

	return nil
}

// GetServer 获取服务器信息
func (s *ServerService) GetServer(ctx context.Context, sessionID string) (*models.ServerInfo, error) {
	// 先从缓存获取
	if serverInfo, err := s.getCachedServerInfo(ctx, sessionID); err == nil {
		return serverInfo, nil
	}

	// 从数据库获取
	serverInfo, err := s.serverRepo.GetServer(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	// 更新缓存
	s.cacheServerInfo(ctx, serverInfo)

	return serverInfo, nil
}

// GetServers 获取服务器列表
func (s *ServerService) GetServers(ctx context.Context, filter *repository.ServerFilter) ([]*models.ServerInfo, error) {
	// 构建分页参数
	pagination := &repository.Pagination{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	if filter.Limit > 0 {
		pagination.Limit = filter.Limit
	}

	// 根据过滤条件获取服务器
	var servers []*models.ServerInfo
	var err error

	if filter.ProjectKey != "" {
		servers, err = s.serverRepo.GetServersByProject(ctx, filter.ProjectKey, pagination)
	} else {
		servers, err = s.serverRepo.GetAllServers(ctx, "", pagination.Offset, pagination.Limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get servers: %w", err)
	}

	// 应用其他过滤条件
	servers = s.applyFilters(servers, filter)

	return servers, nil
}

// RemoveServer 移除服务器
func (s *ServerService) RemoveServer(ctx context.Context, sessionID string) error {
	// 从数据库删除
	if err := s.serverRepo.DeleteServer(ctx, sessionID); err != nil {
		s.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to remove server")
		return fmt.Errorf("failed to remove server: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("server:%s", sessionID)
	if err := s.cacheRepo.Delete(ctx, cacheKey); err != nil {
		s.logger.WithError(err).WithField("session_id", sessionID).Warn("Failed to clear server cache")
	}

	s.logger.WithField("session_id", sessionID).Info("Server removed successfully")
	return nil
}

// GetServersByProject 按项目获取服务器
func (s *ServerService) GetServersByProject(ctx context.Context, projectKey string, pagination *repository.Pagination) ([]*models.ServerInfo, error) {
	servers, err := s.serverRepo.GetServersByProject(ctx, projectKey, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers by project: %w", err)
	}

	return servers, nil
}

// GetOnlineServers 获取在线服务器
func (s *ServerService) GetOnlineServers(ctx context.Context, projectKey string) ([]*models.ServerInfo, error) {
	servers, err := s.serverRepo.GetOnlineServers(ctx, projectKey, s.offlineTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to get online servers: %w", err)
	}

	return servers, nil
}

// GetServerCount 获取服务器数量
func (s *ServerService) GetServerCount(ctx context.Context, projectKey string) (int, error) {
	count, err := s.serverRepo.GetServerCount(ctx, projectKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get server count: %w", err)
	}

	return count, nil
}

// GetProjectStats 获取项目统计信息
func (s *ServerService) GetProjectStats(ctx context.Context, projectKey string) (*repository.ProjectStats, error) {
	// 获取总服务器数
	totalServers, err := s.GetServerCount(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get total server count: %w", err)
	}

	// 获取在线服务器数
	onlineServers, err := s.GetOnlineServers(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get online servers: %w", err)
	}

	// 计算平均资源使用率
	avgCPU, avgMem, avgDisk := s.calculateAverageUsage(onlineServers)

	stats := &repository.ProjectStats{
		TotalServers:   totalServers,
		OnlineServers:  len(onlineServers),
		OfflineServers: totalServers - len(onlineServers),
		AvgCPUUsage:    avgCPU,
		AvgMemoryUsage: avgMem,
		AvgDiskUsage:   avgDisk,
		LastUpdateTime: time.Now(),
	}

	return stats, nil
}

// UpdateServerOnlineStatus 更新服务器在线状态
func (s *ServerService) UpdateServerOnlineStatus(ctx context.Context, sessionID string, online bool) error {
	// 这里可以通过更新last_seen字段来管理在线状态
	if online {
		return s.serverRepo.UpdateLastSeen(ctx, sessionID)
	}
	return nil
}

// MarkServerAsOffline 标记服务器为离线
func (s *ServerService) MarkServerAsOffline(ctx context.Context, timeout time.Duration) error {
	// 获取所有服务器
	servers, err := s.serverRepo.GetAllServers(ctx, "", 0, 1000) // 限制数量避免过大查询
	if err != nil {
		return fmt.Errorf("failed to get servers for offline check: %w", err)
	}

	offlineCount := 0
	for _, server := range servers {
		// 检查最后更新时间
		if time.Since(server.UpdatedAt) > timeout {
			server.IsOnline = false
			// 更新状态
			if err := s.serverRepo.UpdateServer(ctx, server); err != nil {
				s.logger.WithError(err).WithField("session_id", server.SessionID).Warn("Failed to mark server as offline")
				continue
			}
			offlineCount++
		}
	}

	if offlineCount > 0 {
		s.logger.WithField("offline_count", offlineCount).Info("Servers marked as offline")
	}

	return nil
}

// validateSystemInfo 验证系统信息
func (s *ServerService) validateSystemInfo(info *models.SystemInfo) error {
	if info.SessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	if info.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	if info.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	return nil
}

// saveHistoryData 保存历史数据
func (s *ServerService) saveHistoryData(ctx context.Context, info *models.SystemInfo) error {
	historyData := &models.HistoryData{
		SessionID:   info.SessionID,
		Hostname:    info.Hostname,
		ProjectKey:  info.ProjectKey,
		CPUUsage:    info.CPUUsage,
		MemoryUsed:  info.MemInfo.Used,
		MemoryUsage: info.MemInfo.UsagePercent,
		DiskUsed:    info.DiskInfo.Used,
		DiskUsage:   info.DiskInfo.UsagePercent,
		NetworkTx:   info.NetInfo.Tx,
		NetworkRx:   info.NetInfo.Rx,
		GPUUsage:    info.GPUInfo.Usage,
		LoadAvg:     info.LoadAvg,
		Temperature: info.TempInfo.Temp,
		Timestamp:   info.Timestamp,
	}

	return s.historyRepo.SaveHistoryData(ctx, historyData)
}

// cacheServerInfo 缓存服务器信息
func (s *ServerService) cacheServerInfo(ctx context.Context, serverInfo *models.ServerInfo) {
	cacheKey := fmt.Sprintf("server:%s", serverInfo.SessionID)
	ttl := 5 * time.Minute
	s.cacheRepo.Set(ctx, cacheKey, serverInfo, ttl)
}

// getCachedServerInfo 获取缓存的服务器信息
func (s *ServerService) getCachedServerInfo(ctx context.Context, sessionID string) (*models.ServerInfo, error) {
	cacheKey := fmt.Sprintf("server:%s", sessionID)
	var serverInfo models.ServerInfo
	err := s.cacheRepo.Get(ctx, cacheKey, &serverInfo)
	return &serverInfo, err
}

// applyFilters 应用过滤条件
func (s *ServerService) applyFilters(servers []*models.ServerInfo, filter *repository.ServerFilter) []*models.ServerInfo {
	var result []*models.ServerInfo

	for _, server := range servers {
		// 状态过滤
		if filter.Status != "" {
			switch filter.Status {
			case "online":
				if !server.IsOnline {
					continue
				}
			case "offline":
				if server.IsOnline {
					continue
				}
			}
		}

		// 主机名过滤
		if filter.Hostname != "" && server.Hostname != filter.Hostname {
			continue
		}

		// 操作系统过滤
		if filter.OS != "" && server.OS != filter.OS {
			continue
		}

		// 标签过滤（简单实现）
		if len(filter.Tags) > 0 {
			// 这里可以实现更复杂的标签匹配逻辑
			continue
		}

		// 时间范围过滤
		if filter.LastSeenAfter != nil && server.UpdatedAt.Before(*filter.LastSeenAfter) {
			continue
		}
		if filter.LastSeenBefore != nil && server.UpdatedAt.After(*filter.LastSeenBefore) {
			continue
		}

		result = append(result, server)
	}

	return result
}

// calculateAverageUsage 计算平均资源使用率
func (s *ServerService) calculateAverageUsage(servers []*models.ServerInfo) (avgCPU, avgMem, avgDisk float64) {
	if len(servers) == 0 {
		return 0, 0, 0
	}

	var totalCPU, totalMem, totalDisk float64

	for _, server := range servers {
		if server.SystemInfo != nil {
			totalCPU += server.SystemInfo.CPUUsage
			totalMem += server.SystemInfo.MemInfo.UsagePercent
			totalDisk += server.SystemInfo.DiskInfo.UsagePercent
		}
	}

	count := float64(len(servers))
	return totalCPU / count, totalMem / count, totalDisk / count
}
