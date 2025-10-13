package auth

import (
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// AuthServiceFactory 认证服务工厂
type AuthServiceFactory struct {
	repository repository.Repository
	logger     logger.Logger
}

// NewAuthServiceFactory 创建认证服务工厂
func NewAuthServiceFactory(repository repository.Repository, logger logger.Logger) *AuthServiceFactory {
	return &AuthServiceFactory{
		repository: repository,
		logger:     logger,
	}
}

// CreateAuthService 创建认证服务
func (f *AuthServiceFactory) CreateAuthService(config *Config) *AuthService {
	// 尝试访问嵌入的repositories
	switch repo := f.repository.(type) {
	case interface {
		AccessKeyRepository() repository.AccessKeyRepository
		CacheRepository() repository.CacheRepository
	}:
		return NewAuthService(
			repo.AccessKeyRepository(),
			repo.CacheRepository(),
			f.logger,
			config,
		)
	default:
		// 回退：尝试直接类型转换
		if accessKeyRepo, ok := f.repository.(repository.AccessKeyRepository); ok {
			if cacheRepo, ok := f.repository.(repository.CacheRepository); ok {
				return NewAuthService(accessKeyRepo, cacheRepo, f.logger, config)
			}
		}
		// 作为最后的选择，panic或返回nil
		panic("repository does not implement required interfaces for auth service")
	}
}

// CreateAuthServiceWithDependencies 创建带依赖的认证服务
func (f *AuthServiceFactory) CreateAuthServiceWithDependencies(
	accessKeyRepo repository.AccessKeyRepository,
	cacheRepo repository.CacheRepository,
	config *Config,
) *AuthService {
	return NewAuthService(accessKeyRepo, cacheRepo, f.logger, config)
}

// CreateDefaultAuthService 创建使用默认配置的认证服务
func (f *AuthServiceFactory) CreateDefaultAuthService() *AuthService {
	return f.CreateAuthService(DefaultConfig())
}
