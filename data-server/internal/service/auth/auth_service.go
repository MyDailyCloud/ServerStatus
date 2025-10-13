package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// AuthService 认证授权服务
type AuthService struct {
	accessKeyRepo repository.AccessKeyRepository
	cacheRepo     repository.CacheRepository
	logger        logger.Logger
	config        *Config
}

// Config 认证服务配置
type Config struct {
	// 密钥配置
	TokenExpiration time.Duration `yaml:"token_expiration" json:"token_expiration"`

	// 缓存配置
	CacheEnabled bool          `yaml:"cache_enabled" json:"cache_enabled"`
	CacheTTL     time.Duration `yaml:"cache_ttl" json:"cache_ttl"`

	// 安全配置
	MaxFailedAttempts int           `yaml:"max_failed_attempts" json:"max_failed_attempts"`
	LockoutDuration   time.Duration `yaml:"lockout_duration" json:"lockout_duration"`

	// 管理员配置
	AdminAccessKey string `yaml:"admin_access_key" json:"admin_access_key"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		TokenExpiration:   24 * time.Hour,
		CacheEnabled:      true,
		CacheTTL:          15 * time.Minute,
		MaxFailedAttempts: 5,
		LockoutDuration:   15 * time.Minute,
	}
}

// NewAuthService 创建认证授权服务
func NewAuthService(
	accessKeyRepo repository.AccessKeyRepository,
	cacheRepo repository.CacheRepository,
	logger logger.Logger,
	config *Config,
) *AuthService {
	if config == nil {
		config = DefaultConfig()
	}

	return &AuthService{
		accessKeyRepo: accessKeyRepo,
		cacheRepo:     cacheRepo,
		logger:        logger,
		config:        config,
	}
}

// Permission 权限定义
type Permission string

const (
	PermissionRead      Permission = "read"      // 读取权限
	PermissionWrite     Permission = "write"     // 写入权限
	PermissionDelete    Permission = "delete"    // 删除权限
	PermissionExport    Permission = "export"    // 导出权限
	PermissionAdmin     Permission = "admin"     // 管理权限
	PermissionWebSocket Permission = "websocket" // WebSocket权限
)

// AuthRequest 认证请求
type AuthRequest struct {
	AccessKey  string      `json:"access_key"`
	ProjectKey string      `json:"project_key,omitempty"`
	Permission Permission  `json:"permission,omitempty"`
	ClientInfo *ClientInfo `json:"client_info,omitempty"`
}

// ClientInfo 客户端信息
type ClientInfo struct {
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
}

// AuthResult 认证结果
type AuthResult struct {
	Success     bool          `json:"success"`
	ProjectKey  string        `json:"project_key,omitempty"`
	Permissions []Permission  `json:"permissions,omitempty"`
	Message     string        `json:"message,omitempty"`
	TTL         time.Duration `json:"ttl,omitempty"`
}

// ValidateAccessKey 验证访问密钥
func (s *AuthService) ValidateAccessKey(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
	if req.AccessKey == "" {
		return &AuthResult{
			Success: false,
			Message: "access key is required",
		}, nil
	}

	// 检查管理员密钥
	if s.config.AdminAccessKey != "" && req.AccessKey == s.config.AdminAccessKey {
		return &AuthResult{
			Success:     true,
			ProjectKey:  "*",
			Permissions: []Permission{PermissionRead, PermissionWrite, PermissionDelete, PermissionExport, PermissionAdmin, PermissionWebSocket},
			Message:     "admin access granted",
			TTL:         s.config.TokenExpiration,
		}, nil
	}

	// 检查缓存
	if s.config.CacheEnabled {
		if cached, err := s.getCachedAuthResult(ctx, req.AccessKey); err == nil && cached != nil {
			s.logger.WithField("access_key", req.AccessKey[:8]+"...").Debug("Access key validated from cache")
			return cached, nil
		}
	}

	// 检查是否被锁定
	if isLocked, err := s.isAccessKeyLocked(ctx, req.AccessKey); err == nil && isLocked {
		return &AuthResult{
			Success: false,
			Message: "access key is temporarily locked due to failed attempts",
		}, nil
	}

	// 使用现有的repository接口验证访问密钥
	projectKey, err := s.accessKeyRepo.ValidateAccessKey(ctx, req.AccessKey)
	if err != nil {
		s.recordFailedAttempt(ctx, req.AccessKey)
		return &AuthResult{
			Success: false,
			Message: "invalid access key",
		}, nil
	}

	// 检查项目密钥
	if req.ProjectKey != "" && projectKey != "*" && projectKey != req.ProjectKey {
		return &AuthResult{
			Success: false,
			Message: "access key does not have permission for this project",
		}, nil
	}

	// 获取默认权限
	permissions := s.GetDefaultPermissions()

	// 构建认证结果
	result := &AuthResult{
		Success:     true,
		ProjectKey:  projectKey,
		Permissions: permissions,
		TTL:         s.config.TokenExpiration,
	}

	// 清除失败尝试记录
	s.clearFailedAttempts(ctx, req.AccessKey)

	// 缓存认证结果
	if s.config.CacheEnabled {
		if err := s.cacheAuthResult(ctx, req.AccessKey, result); err != nil {
			s.logger.WithError(err).Warn("Failed to cache auth result")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"access_key": req.AccessKey[:8] + "...",
		"project":    projectKey,
		"success":    true,
	}).Info("Access key validated successfully")

	return result, nil
}

// GenerateAccessKey 生成访问密钥
func (s *AuthService) GenerateAccessKey(ctx context.Context, serverKey, projectKey string) (string, error) {
	accessKey, err := s.accessKeyRepo.GenerateAccessKey(ctx, serverKey, projectKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate access key: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"server_key": serverKey,
		"project":    projectKey,
		"access_key": accessKey[:8] + "...",
	}).Info("Access key generated successfully")

	return accessKey, nil
}

// SaveAccessKey 保存访问密钥
func (s *AuthService) SaveAccessKey(ctx context.Context, cacheKey, accessKey string) error {
	if err := s.accessKeyRepo.SaveAccessKey(ctx, cacheKey, accessKey); err != nil {
		return fmt.Errorf("failed to save access key: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"cache_key":  cacheKey,
		"access_key": accessKey[:8] + "...",
	}).Debug("Access key saved to cache")

	return nil
}

// DeleteAccessKey 删除访问密钥
func (s *AuthService) DeleteAccessKey(ctx context.Context, accessKey string) error {
	if err := s.accessKeyRepo.DeleteAccessKey(ctx, accessKey); err != nil {
		return fmt.Errorf("failed to delete access key: %w", err)
	}

	// 清除相关缓存
	s.clearAuthCache(ctx, accessKey)

	s.logger.WithField("access_key", accessKey[:8]+"...").Info("Access key deleted successfully")
	return nil
}

// CleanupExpiredKeys 清理过期密钥
func (s *AuthService) CleanupExpiredKeys(ctx context.Context) error {
	ttl := s.config.TokenExpiration
	if err := s.accessKeyRepo.CleanupExpiredKeys(ctx, ttl); err != nil {
		return fmt.Errorf("failed to cleanup expired keys: %w", err)
	}

	s.logger.WithField("ttl", ttl).Info("Expired access keys cleaned up successfully")
	return nil
}

// getCachedAuthResult 从缓存获取认证结果
func (s *AuthService) getCachedAuthResult(ctx context.Context, accessKey string) (*AuthResult, error) {
	if !s.config.CacheEnabled {
		return nil, fmt.Errorf("cache is disabled")
	}

	cacheKey := fmt.Sprintf("auth:result:%s", accessKey)

	var result AuthResult
	if err := s.cacheRepo.Get(ctx, cacheKey, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// cacheAuthResult 缓存认证结果
func (s *AuthService) cacheAuthResult(ctx context.Context, accessKey string, result *AuthResult) error {
	if !s.config.CacheEnabled {
		return nil
	}

	cacheKey := fmt.Sprintf("auth:result:%s", accessKey)
	return s.cacheRepo.Set(ctx, cacheKey, result, s.config.CacheTTL)
}

// clearAuthCache 清除认证缓存
func (s *AuthService) clearAuthCache(ctx context.Context, accessKey string) {
	if !s.config.CacheEnabled {
		return
	}

	cacheKey := fmt.Sprintf("auth:result:%s", accessKey)
	if err := s.cacheRepo.Delete(ctx, cacheKey); err != nil {
		s.logger.WithError(err).Warn("Failed to clear auth cache")
	}
}

// isAccessKeyLocked 检查访问密钥是否被锁定
func (s *AuthService) isAccessKeyLocked(ctx context.Context, accessKey string) (bool, error) {
	if !s.config.CacheEnabled {
		return false, nil
	}

	cacheKey := fmt.Sprintf("auth:lock:%s", accessKey)
	var locked bool
	if err := s.cacheRepo.Get(ctx, cacheKey, &locked); err != nil {
		return false, nil
	}

	return locked, nil
}

// recordFailedAttempt 记录失败尝试
func (s *AuthService) recordFailedAttempt(ctx context.Context, accessKey string) {
	if !s.config.CacheEnabled {
		return
	}

	cacheKey := fmt.Sprintf("auth:failed:%s", accessKey)

	// 获取当前失败次数
	var failedCount int
	if err := s.cacheRepo.Get(ctx, cacheKey, &failedCount); err != nil {
		failedCount = 0
	}

	failedCount++

	// 更新失败次数
	if err := s.cacheRepo.Set(ctx, cacheKey, failedCount, s.config.LockoutDuration); err != nil {
		s.logger.WithError(err).Warn("Failed to record failed attempt")
		return
	}

	// 如果超过最大失败次数，锁定访问密钥
	if failedCount >= s.config.MaxFailedAttempts {
		lockKey := fmt.Sprintf("auth:lock:%s", accessKey)
		if err := s.cacheRepo.Set(ctx, lockKey, true, s.config.LockoutDuration); err != nil {
			s.logger.WithError(err).Warn("Failed to lock access key")
		}

		s.logger.WithField("access_key", accessKey[:8]+"...").Warn("Access key locked due to too many failed attempts")
	}
}

// clearFailedAttempts 清除失败尝试记录
func (s *AuthService) clearFailedAttempts(ctx context.Context, accessKey string) {
	if !s.config.CacheEnabled {
		return
	}

	cacheKey := fmt.Sprintf("auth:failed:%s", accessKey)
	if err := s.cacheRepo.Delete(ctx, cacheKey); err != nil {
		s.logger.WithError(err).Warn("Failed to clear failed attempts")
	}

	lockKey := fmt.Sprintf("auth:lock:%s", accessKey)
	if err := s.cacheRepo.Delete(ctx, lockKey); err != nil {
		s.logger.WithError(err).Warn("Failed to clear access key lock")
	}
}

// ValidatePermissionString 验证权限字符串
func (s *AuthService) ValidatePermissionString(permission string) bool {
	switch Permission(permission) {
	case PermissionRead, PermissionWrite, PermissionDelete, PermissionExport, PermissionAdmin, PermissionWebSocket:
		return true
	default:
		return false
	}
}

// GetDefaultPermissions 获取默认权限
func (s *AuthService) GetDefaultPermissions() []Permission {
	return []Permission{PermissionRead, PermissionWebSocket}
}

// GetAllPermissions 获取所有权限列表
func (s *AuthService) GetAllPermissions() []Permission {
	return []Permission{
		PermissionRead,
		PermissionWrite,
		PermissionDelete,
		PermissionExport,
		PermissionAdmin,
		PermissionWebSocket,
	}
}

// HasPermission 检查权限集合中是否包含指定权限
func (s *AuthService) HasPermission(permissions []Permission, required Permission) bool {
	for _, p := range permissions {
		if p == required || p == PermissionAdmin {
			return true
		}
	}
	return false
}

// GetAuthStats 获取认证统计信息
func (s *AuthService) GetAuthStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 基本统计信息
	stats["cache_enabled"] = s.config.CacheEnabled
	stats["token_expiration"] = s.config.TokenExpiration.String()
	stats["max_failed_attempts"] = s.config.MaxFailedAttempts
	stats["lockout_duration"] = s.config.LockoutDuration.String()

	// 如果启用缓存，获取缓存统计信息
	if s.config.CacheEnabled && s.cacheRepo != nil {
		if cacheStats, err := s.cacheRepo.GetStats(ctx); err == nil {
			stats["cache_stats"] = cacheStats
		}
	}

	return stats, nil
}
