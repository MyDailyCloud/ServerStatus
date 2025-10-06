package health

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"   // 健康
	StatusUnhealthy HealthStatus = "unhealthy" // 不健康
	StatusDegraded  HealthStatus = "degraded"  // 降级服务
)

// ComponentStatus 组件状态
type ComponentStatus struct {
	Name     string                 `json:"name"`
	Status   HealthStatus           `json:"status"`
	Message  string                 `json:"message,omitempty"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    HealthStatus              `json:"status"`
	Timestamp time.Time                 `json:"timestamp"`
	Duration  time.Duration             `json:"duration"`
	Components map[string]*ComponentStatus `json:"components"`
	Version   string                    `json:"version"`
	Uptime    time.Duration             `json:"uptime"`
	Memory    *MemoryInfo               `json:"memory,omitempty"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`       // 已分配的堆内存 (bytes)
	TotalAlloc uint64 `json:"total_alloc"` // 累计分配的内存 (bytes)
	Sys        uint64 `json:"sys"`         // 从系统获取的内存 (bytes)
	NumGC      uint32 `json:"num_gc"`      // GC运行次数
}

// CheckRequest 健康检查请求
type CheckRequest struct {
	IncludeSystem bool `json:"include_system,omitempty"` // 包含系统信息
	IncludeMemory bool `json:"include_memory,omitempty"` // 包含内存信息
	CheckAll      bool `json:"check_all,omitempty"`      // 检查所有组件
}

// CheckResult 检查结果
type CheckResult struct {
	Healthy bool          `json:"healthy"`
	Status  HealthStatus  `json:"status"`
	Message string        `json:"message,omitempty"`
	Duration time.Duration `json:"duration"`
}

// Config 健康检查服务配置
type Config struct {
	// 检查配置
	Enabled        bool          `yaml:"enabled" json:"enabled"`
	CheckInterval  time.Duration `yaml:"check_interval" json:"check_interval"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`

	// 阈值配置
	MemoryThreshold    uint64        `yaml:"memory_threshold" json:"memory_threshold"`    // 内存阈值 (MB)
	ResponseThreshold  time.Duration `yaml:"response_threshold" json:"response_threshold"`  // 响应时间阈值

	// 系统信息
	Version string `yaml:"version" json:"version"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:           true,
		CheckInterval:     30 * time.Second,
		Timeout:           5 * time.Second,
		MemoryThreshold:   512, // 512MB
		ResponseThreshold: 100 * time.Millisecond,
		Version:           "2.1.0",
	}
}

// HealthService 健康检查服务
type HealthService struct {
	serverRepo    repository.ServerRepository
	cacheRepo     repository.CacheRepository
	accessKeyRepo repository.AccessKeyRepository
	logger        logger.Logger
	config        *Config
	startTime     time.Time
}

// NewHealthService 创建健康检查服务
func NewHealthService(
	serverRepo repository.ServerRepository,
	cacheRepo repository.CacheRepository,
	accessKeyRepo repository.AccessKeyRepository,
	logger logger.Logger,
	config *Config,
) *HealthService {
	if config == nil {
		config = DefaultConfig()
	}

	return &HealthService{
		serverRepo:    serverRepo,
		cacheRepo:     cacheRepo,
		accessKeyRepo: accessKeyRepo,
		logger:        logger,
		config:        config,
		startTime:     time.Now(),
	}
}

// CheckHealth 执行健康检查
func (s *HealthService) CheckHealth(ctx context.Context, req *CheckRequest) (*HealthResponse, error) {
	if req == nil {
		req = &CheckRequest{
			IncludeSystem: true,
			IncludeMemory: true,
			CheckAll:      true,
		}
	}

	startTime := time.Now()
	response := &HealthResponse{
		Timestamp:  startTime,
		Components: make(map[string]*ComponentStatus),
		Version:    s.config.Version,
		Uptime:     time.Since(s.startTime),
	}

	// 检查各个组件
	overallHealthy := true

	// 检查数据库连接
	dbStatus := s.checkDatabase(ctx)
	response.Components["database"] = dbStatus
	if dbStatus.Status != StatusHealthy {
		overallHealthy = false
	}

	// 检查缓存连接
	cacheStatus := s.checkCache(ctx)
	response.Components["cache"] = cacheStatus
	if cacheStatus.Status != StatusHealthy {
		overallHealthy = false
	}

	// 检查访问密钥服务
	authStatus := s.checkAuthService(ctx)
	response.Components["auth"] = authStatus
	if authStatus.Status != StatusHealthy {
		overallHealthy = false
	}

	// 检查系统状态
	if req.IncludeSystem || req.CheckAll {
		systemStatus := s.checkSystem(ctx)
		response.Components["system"] = systemStatus
		if systemStatus.Status != StatusHealthy {
			overallHealthy = false
		}
	}

	// 检查内存状态
	if req.IncludeMemory || req.CheckAll {
		memoryStatus := s.checkMemory(ctx)
		response.Components["memory"] = memoryStatus
		response.Memory = memoryStatus.Metadata["memory_info"].(*MemoryInfo)
		if memoryStatus.Status != StatusHealthy {
			overallHealthy = false
		}
	}

	// 确定整体状态
	response.Duration = time.Since(startTime)
	if overallHealthy {
		response.Status = StatusHealthy
	} else {
		// 检查是否有任何组件是健康的
		hasHealthy := false
		for _, component := range response.Components {
			if component.Status == StatusHealthy {
				hasHealthy = true
				break
			}
		}

		if hasHealthy {
			response.Status = StatusDegraded
		} else {
			response.Status = StatusUnhealthy
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"status":      response.Status,
		"duration":    response.Duration,
		"components":  len(response.Components),
	}).Debug("Health check completed")

	return response, nil
}

// checkDatabase 检查数据库连接
func (s *HealthService) checkDatabase(ctx context.Context) *ComponentStatus {
	startTime := time.Now()

	// 使用repository的Ping方法检查连接
	if err := s.serverRepo.Ping(); err != nil {
		return &ComponentStatus{
			Name:     "database",
			Status:   StatusUnhealthy,
			Message:  fmt.Sprintf("database connection failed: %v", err),
			Duration: time.Since(startTime),
		}
	}

	return &ComponentStatus{
		Name:     "database",
		Status:   StatusHealthy,
		Message:  "database connection is healthy",
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"type": "sqlite",
		},
	}
}

// checkCache 检查缓存连接
func (s *HealthService) checkCache(ctx context.Context) *ComponentStatus {
	startTime := time.Now()

	// 检查缓存服务是否连接
	if !s.cacheRepo.IsConnected() {
		return &ComponentStatus{
			Name:     "cache",
			Status:   StatusDegraded,
			Message:  "cache service is not connected",
			Duration: time.Since(startTime),
		}
	}

	// 尝试设置和获取测试值
	testKey := fmt.Sprintf("health:check:%d", time.Now().Unix())
	testValue := "ok"

	if err := s.cacheRepo.Set(ctx, testKey, testValue, 10*time.Second); err != nil {
		return &ComponentStatus{
			Name:     "cache",
			Status:   StatusDegraded,
			Message:  fmt.Sprintf("cache write failed: %v", err),
			Duration: time.Since(startTime),
		}
	}

	var result string
	if err := s.cacheRepo.Get(ctx, testKey, &result); err != nil || result != testValue {
		return &ComponentStatus{
			Name:     "cache",
			Status:   StatusDegraded,
			Message:  "cache read/write verification failed",
			Duration: time.Since(startTime),
		}
	}

	// 清理测试数据
	s.cacheRepo.Delete(ctx, testKey)

	return &ComponentStatus{
		Name:     "cache",
		Status:   StatusHealthy,
		Message:  "cache service is healthy",
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"type": s.cacheRepo.GetType(),
		},
	}
}

// checkAuthService 检查认证服务
func (s *HealthService) checkAuthService(ctx context.Context) *ComponentStatus {
	startTime := time.Now()

	// 通过验证一个测试密钥来检查认证服务是否正常工作
	// 这里我们尝试生成一个密钥来验证服务功能
	_, err := s.accessKeyRepo.GenerateAccessKey(ctx, "test-key", "test-project")
	if err != nil {
		return &ComponentStatus{
			Name:     "auth",
			Status:   StatusUnhealthy,
			Message:  fmt.Sprintf("auth service validation failed: %v", err),
			Duration: time.Since(startTime),
		}
	}

	return &ComponentStatus{
		Name:     "auth",
		Status:   StatusHealthy,
		Message:  "auth service is healthy",
		Duration: time.Since(startTime),
	}
}

// checkSystem 检查系统状态
func (s *HealthService) checkSystem(ctx context.Context) *ComponentStatus {
	startTime := time.Now()

	metadata := map[string]interface{}{
		"go_version":   runtime.Version(),
		"goroutines":   runtime.NumGoroutine(),
		"cpu_count":    runtime.NumCPU(),
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
	}

	// 检查Goroutine数量是否过高
	goroutineCount := runtime.NumGoroutine()
	if goroutineCount > 1000 {
		return &ComponentStatus{
			Name:     "system",
			Status:   StatusDegraded,
			Message:  fmt.Sprintf("high goroutine count: %d", goroutineCount),
			Duration: time.Since(startTime),
			Metadata: metadata,
		}
	}

	return &ComponentStatus{
		Name:     "system",
		Status:   StatusHealthy,
		Message:  "system is healthy",
		Duration: time.Since(startTime),
		Metadata: metadata,
	}
}

// checkMemory 检查内存状态
func (s *HealthService) checkMemory(ctx context.Context) *ComponentStatus {
	startTime := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryInfo := &MemoryInfo{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}

	// 转换为MB
	allocMB := m.Alloc / 1024 / 1024
	thresholdMB := s.config.MemoryThreshold

	status := StatusHealthy
	message := fmt.Sprintf("memory usage: %d MB", allocMB)

	if allocMB > thresholdMB {
		status = StatusDegraded
		message = fmt.Sprintf("high memory usage: %d MB (threshold: %d MB)", allocMB, thresholdMB)
	}

	return &ComponentStatus{
		Name:     "memory",
		Status:   status,
		Message:  message,
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"memory_info": memoryInfo,
			"alloc_mb":    allocMB,
			"threshold_mb": thresholdMB,
		},
	}
}

// IsHealthy 快速健康检查
func (s *HealthService) IsHealthy(ctx context.Context) bool {
	req := &CheckRequest{
		IncludeSystem: false,
		IncludeMemory: false,
		CheckAll:      false,
	}

	response, err := s.CheckHealth(ctx, req)
	if err != nil {
		return false
	}

	return response.Status == StatusHealthy
}

// GetServiceInfo 获取服务信息
func (s *HealthService) GetServiceInfo(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"name":       "ServerStatus Data Server",
		"version":    s.config.Version,
		"uptime":     time.Since(s.startTime).String(),
		"start_time": s.startTime,
		"go_version": runtime.Version(),
		"build_time": "unknown", // 可以在构建时设置
	}
}

// GetComponentStatus 获取特定组件状态
func (s *HealthService) GetComponentStatus(ctx context.Context, componentName string) (*ComponentStatus, error) {
	switch componentName {
	case "database":
		return s.checkDatabase(ctx), nil
	case "cache":
		return s.checkCache(ctx), nil
	case "auth":
		return s.checkAuthService(ctx), nil
	case "system":
		return s.checkSystem(ctx), nil
	case "memory":
		return s.checkMemory(ctx), nil
	default:
		return nil, fmt.Errorf("unknown component: %s", componentName)
	}
}

// GetConfig 获取服务配置
func (s *HealthService) GetConfig() *Config {
	return s.config
}

// UpdateConfig 更新服务配置
func (s *HealthService) UpdateConfig(config *Config) {
	if config != nil {
		s.config = config
		s.logger.WithField("config", "updated").Info("Health service configuration updated")
	}
}