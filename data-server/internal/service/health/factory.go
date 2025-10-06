package health

import (
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Factory 健康检查服务工厂
type Factory struct {
	serverRepo    repository.ServerRepository
	cacheRepo     repository.CacheRepository
	accessKeyRepo repository.AccessKeyRepository
	logger        logger.Logger
}

// NewFactory 创建健康检查服务工厂
func NewFactory(
	serverRepo repository.ServerRepository,
	cacheRepo repository.CacheRepository,
	accessKeyRepo repository.AccessKeyRepository,
	logger logger.Logger,
) *Factory {
	return &Factory{
		serverRepo:    serverRepo,
		cacheRepo:     cacheRepo,
		accessKeyRepo: accessKeyRepo,
		logger:        logger,
	}
}

// CreateService 创建健康检查服务
func (f *Factory) CreateService(config *Config) *HealthService {
	return NewHealthService(
		f.serverRepo,
		f.cacheRepo,
		f.accessKeyRepo,
		f.logger,
		config,
	)
}

// CreateDefaultService 创建默认配置的健康检查服务
func (f *Factory) CreateDefaultService() *HealthService {
	return f.CreateService(DefaultConfig())
}

// CreateServiceWithDependencies 使用依赖注入创建健康检查服务
func CreateServiceWithDependencies(
	serverRepo repository.ServerRepository,
	cacheRepo repository.CacheRepository,
	accessKeyRepo repository.AccessKeyRepository,
	logger logger.Logger,
	config *Config,
) *HealthService {
	return NewHealthService(serverRepo, cacheRepo, accessKeyRepo, logger, config)
}

// CreateServiceFromInterfaces 从接口创建健康检查服务
func CreateServiceFromInterfaces(
	serverRepo interface{},
	cacheRepo interface{},
	accessKeyRepo interface{},
	loggerInterface interface{},
	config *Config,
) *HealthService {
	// 类型断言，确保传入的是正确的接口类型
	srvRepo, ok := serverRepo.(repository.ServerRepository)
	if !ok {
		panic("serverRepo must implement repository.ServerRepository interface")
	}

	cRepo, ok := cacheRepo.(repository.CacheRepository)
	if !ok {
		panic("cacheRepo must implement repository.CacheRepository interface")
	}

	akRepo, ok := accessKeyRepo.(repository.AccessKeyRepository)
	if !ok {
		panic("accessKeyRepo must implement repository.AccessKeyRepository interface")
	}

	log, ok := loggerInterface.(logger.Logger)
	if !ok {
		panic("logger must implement logger.Logger interface")
	}

	return NewHealthService(srvRepo, cRepo, akRepo, log, config)
}