package server

import (
	"context"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// ServerServiceFactory 服务器服务工厂
type ServerServiceFactory struct {
	logger logger.Logger
}

// NewServerServiceFactory 创建服务器服务工厂
func NewServerServiceFactory(logger logger.Logger) *ServerServiceFactory {
	return &ServerServiceFactory{
		logger: logger,
	}
}

// CreateServerService 创建服务器服务实例
func (f *ServerServiceFactory) CreateServerService(
	serverRepo repository.ServerRepository,
	historyRepo repository.HistoryRepository,
	cacheRepo repository.CacheRepository,
) (*ServerService, error) {
	return NewServerService(serverRepo, historyRepo, cacheRepo, f.logger), nil
}

// ServerServiceConfig 服务器服务配置
type ServerServiceConfig struct {
	OfflineTimeout string `yaml:"offline_timeout" json:"offline_timeout"`
	CacheTTL       string `yaml:"cache_ttl" json:"cache_ttl"`
	BatchSize      int    `yaml:"batch_size" json:"batch_size"`
}

// DefaultServerServiceConfig 返回默认服务器服务配置
func DefaultServerServiceConfig() *ServerServiceConfig {
	return &ServerServiceConfig{
		OfflineTimeout: "5m",
		CacheTTL:       "5m",
		BatchSize:      100,
	}
}

// ServerServiceManager 服务器服务管理器
type ServerServiceManager struct {
	serverService *ServerService
	logger        logger.Logger
}

// NewServerServiceManager 创建服务器服务管理器
func NewServerServiceManager(
	serverRepo repository.ServerRepository,
	historyRepo repository.HistoryRepository,
	cacheRepo repository.CacheRepository,
	logger logger.Logger,
) *ServerServiceManager {
	serverService := NewServerService(serverRepo, historyRepo, cacheRepo, logger)

	return &ServerServiceManager{
		serverService: serverService,
		logger:        logger,
	}
}

// GetServerService 获取服务器服务
func (m *ServerServiceManager) GetServerService() *ServerService {
	return m.serverService
}

// CleanupExpiredData 清理过期数据
func (m *ServerServiceManager) CleanupExpiredData(ctx context.Context) error {
	m.logger.Info("Starting cleanup of expired server data")

	// 标记离线服务器
	if err := m.serverService.MarkServerAsOffline(ctx, 30*time.Minute); err != nil {
		m.logger.WithError(err).Error("Failed to mark offline servers")
		return err
	}

	// 清理过期历史数据
	// 这里可以添加更多的清理逻辑

	m.logger.Info("Cleanup of expired server data completed")
	return nil
}

// GetServiceHealth 获取服务健康状态
func (m *ServerServiceManager) GetServiceHealth(ctx context.Context) map[string]interface{} {
	health := make(map[string]interface{})

	// 检查Repository连接
	if err := m.serverService.serverRepo.Ping(); err != nil {
		health["server_repository"] = "unhealthy"
		health["server_repository_error"] = err.Error()
	} else {
		health["server_repository"] = "healthy"
	}

	if err := m.serverService.historyRepo.Ping(); err != nil {
		health["history_repository"] = "unhealthy"
		health["history_repository_error"] = err.Error()
	} else {
		health["history_repository"] = "healthy"
	}

	// 检查缓存连接
	if m.serverService.cacheRepo.IsConnected() {
		health["cache_repository"] = "healthy"
	} else {
		health["cache_repository"] = "unhealthy"
	}

	// 检查服务状态
	health["service"] = "healthy"

	return health
}
