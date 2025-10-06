package websocket

import (
	"context"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Factory WebSocket服务工厂
type Factory struct {
	serverRepo    repository.ServerRepository
	cacheRepo     repository.CacheRepository
	accessKeyRepo repository.AccessKeyRepository
	logger        logger.Logger
}

// NewFactory 创建WebSocket服务工厂
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

// CreateService 创建WebSocket服务
func (f *Factory) CreateService(config *Config) *WebSocketService {
	// 创建认证服务适配器
	authService := &authServiceAdapter{
		authService: f.accessKeyRepo,
	}

	return NewWebSocketService(
		f.serverRepo,
		f.cacheRepo,
		f.accessKeyRepo,
		authService,
		f.logger,
		config,
	)
}

// CreateDefaultService 创建默认配置的WebSocket服务
func (f *Factory) CreateDefaultService() *WebSocketService {
	return f.CreateService(DefaultConfig())
}

// CreateServiceWithDependencies 使用依赖注入创建WebSocket服务
func CreateServiceWithDependencies(
	serverRepo repository.ServerRepository,
	cacheRepo repository.CacheRepository,
	accessKeyRepo repository.AccessKeyRepository,
	authService AuthService,
	logger logger.Logger,
	config *Config,
) *WebSocketService {
	return NewWebSocketService(serverRepo, cacheRepo, accessKeyRepo, authService, logger, config)
}

// CreateServiceFromInterfaces 从接口创建WebSocket服务
func CreateServiceFromInterfaces(
	serverRepo interface{},
	cacheRepo interface{},
	accessKeyRepo interface{},
	authService interface{},
	loggerInterface interface{},
	config *Config,
) *WebSocketService {
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

	auth, ok := authService.(AuthService)
	if !ok {
		panic("authService must implement AuthService interface")
	}

	log, ok := loggerInterface.(logger.Logger)
	if !ok {
		panic("logger must implement logger.Logger interface")
	}

	return NewWebSocketService(srvRepo, cRepo, akRepo, auth, log, config)
}

// authServiceAdapter 认证服务适配器
// 将AccessKeyRepository适配为AuthService接口
type authServiceAdapter struct {
	authService interface {
		ValidateAccessKey(ctx context.Context, accessKey string) (string, error)
	}
}

// ValidateAccessKey 实现AuthService接口
func (a *authServiceAdapter) ValidateAccessKey(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
	projectKey, err := a.authService.ValidateAccessKey(ctx, req.AccessKey)
	if err != nil {
		return &AuthResult{
			Success: false,
			Message: "invalid access key",
		}, nil
	}

	return &AuthResult{
		Success:     true,
		ProjectKey:  projectKey,
		Permissions: []string{"read", "write", "websocket"}, // 默认权限
		Message:     "access key validated successfully",
	}, nil
}