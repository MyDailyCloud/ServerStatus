package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
	"github.com/kanshan/ServerStatus/data-server/pkg/utils"
	"github.com/go-redis/redis/v8"
)

// RedisAccessKeyRepository Redis访问密钥仓库实现
type RedisAccessKeyRepository struct {
	client *redis.Client
	logger logger.Logger
	prefix string
}

// NewRedisAccessKeyRepository 创建Redis访问密钥仓库
func NewRedisAccessKeyRepository(client *redis.Client, logger logger.Logger, prefix string) repository.AccessKeyRepository {
	return &RedisAccessKeyRepository{
		client: client,
		logger: logger,
		prefix: prefix + ":access_key",
	}
}

// buildKey 构建完整的键名
func (r *RedisAccessKeyRepository) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

// AccessKeyInfo 访问密钥信息
type AccessKeyInfo struct {
	CacheKey    string    `json:"cache_key"`
	ProjectKey  string    `json:"project_key"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsActive    bool      `json:"is_active"`
}

// SaveAccessKey 保存访问密钥
func (r *RedisAccessKeyRepository) SaveAccessKey(ctx context.Context, cacheKey, accessKey string) error {
	fullKey := r.buildKey(accessKey)

	// 解析cacheKey获取projectKey
	projectKey := r.extractProjectKey(cacheKey)
	expiresAt := time.Now().Add(24 * time.Hour) // 默认24小时过期

	// 构建访问密钥信息
	info := AccessKeyInfo{
		CacheKey:   cacheKey,
		ProjectKey: projectKey,
		CreatedAt:  time.Now(),
		ExpiresAt:  expiresAt,
		IsActive:   true,
	}

	// 序列化信息
	data, err := json.Marshal(info)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"access_key": accessKey[:8] + "...",
			"cache_key":  cacheKey,
		}).Error("Failed to marshal access key info")
		return fmt.Errorf("failed to marshal access key info: %w", err)
	}

	// 设置访问密钥信息，使用TTL自动过期
	ttl := expiresAt.Sub(time.Now())
	if err := r.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"access_key": accessKey[:8] + "...",
			"cache_key":  cacheKey,
		}).Error("Failed to save access key")
		return fmt.Errorf("failed to save access key: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"access_key": accessKey[:8] + "...",
		"project_key": projectKey,
		"ttl":        ttl,
	}).Info("Access key saved successfully")

	return nil
}

// GetAccessKey 获取访问密钥
func (r *RedisAccessKeyRepository) GetAccessKey(ctx context.Context, accessKey string) (string, error) {
	fullKey := r.buildKey(accessKey)

	// 获取访问密钥信息
	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("access key not found: %s", accessKey)
		}
		r.logger.WithError(err).WithField("access_key", accessKey[:8]+"...").Error("Failed to get access key")
		return "", fmt.Errorf("failed to get access key: %w", err)
	}

	// 反序列化信息
	var info AccessKeyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		r.logger.WithError(err).WithField("access_key", accessKey[:8]+"...").Error("Failed to unmarshal access key info")
		return "", fmt.Errorf("failed to unmarshal access key info: %w", err)
	}

	// 检查是否有效
	if !info.IsActive {
		return "", fmt.Errorf("access key is inactive: %s", accessKey)
	}

	// 检查是否过期
	if info.ExpiresAt.Before(time.Now()) {
		return "", fmt.Errorf("access key expired: %s", accessKey)
	}

	return info.ProjectKey, nil
}

// DeleteAccessKey 删除访问密钥
func (r *RedisAccessKeyRepository) DeleteAccessKey(ctx context.Context, accessKey string) error {
	fullKey := r.buildKey(accessKey)

	if err := r.client.Del(ctx, fullKey).Err(); err != nil {
		r.logger.WithError(err).WithField("access_key", accessKey[:8]+"...").Error("Failed to delete access key")
		return fmt.Errorf("failed to delete access key: %w", err)
	}

	r.logger.WithField("access_key", accessKey[:8]+"...").Info("Access key deleted successfully")
	return nil
}

// GenerateAccessKey 生成访问密钥
func (r *RedisAccessKeyRepository) GenerateAccessKey(ctx context.Context, serverKey, projectKey string) (string, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("%s:%s", serverKey, projectKey)

	// 生成访问密钥
	accessKey := utils.GenerateAccessKey(serverKey, projectKey)

	// 保存访问密钥
	if err := r.SaveAccessKey(ctx, cacheKey, accessKey); err != nil {
		return "", fmt.Errorf("failed to save generated access key: %w", err)
	}

	return accessKey, nil
}

// ValidateAccessKey 验证访问密钥
func (r *RedisAccessKeyRepository) ValidateAccessKey(ctx context.Context, accessKey string) (string, error) {
	projectKey, err := r.GetAccessKey(ctx, accessKey)
	if err != nil {
		return "", err
	}

	return projectKey, nil
}

// CleanupExpiredKeys 清理过期密钥
func (r *RedisAccessKeyRepository) CleanupExpiredKeys(ctx context.Context, ttl time.Duration) error {
	// Redis会自动清理过期的键，这里主要是清理逻辑上过期的密钥
	pattern := r.buildKey("*")

	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	cleanedCount := 0

	for iter.Next(ctx) {
		key := iter.Val()

		// 获取密钥信息
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue // 键已过期或不存在
		}

		var info AccessKeyInfo
		if err := json.Unmarshal([]byte(data), &info); err != nil {
			// 删除无效的键
			r.client.Del(ctx, key)
			cleanedCount++
			continue
		}

		// 检查是否过期或无效
		if !info.IsActive || info.ExpiresAt.Before(time.Now()) {
			r.client.Del(ctx, key)
			cleanedCount++
		}
	}

	if err := iter.Err(); err != nil {
		r.logger.WithError(err).Error("Failed to scan access keys during cleanup")
		return fmt.Errorf("failed to scan access keys during cleanup: %w", err)
	}

	if cleanedCount > 0 {
		r.logger.WithField("cleaned_count", cleanedCount).Info("Expired access keys cleaned up")
	}

	return nil
}

// extractProjectKey 从缓存键中提取项目键
func (r *RedisAccessKeyRepository) extractProjectKey(cacheKey string) string {
	// 简单的解析逻辑，实际可能需要更复杂的处理
	parts := utils.SplitString(cacheKey, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "default"
}

// GetActiveKeys 获取活跃的访问密钥
func (r *RedisAccessKeyRepository) GetActiveKeys(ctx context.Context, projectKey string) ([]map[string]interface{}, error) {
	pattern := r.buildKey("*")

	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []map[string]interface{}

	for iter.Next(ctx) {
		fullKey := iter.Val()

		// 获取密钥信息
		data, err := r.client.Get(ctx, fullKey).Result()
		if err != nil {
			continue // 键已过期或不存在
		}

		var info AccessKeyInfo
		if err := json.Unmarshal([]byte(data), &info); err != nil {
			continue
		}

		// 过滤项目键
		if projectKey != "" && info.ProjectKey != projectKey {
			continue
		}

		// 检查是否有效且未过期
		if !info.IsActive || info.ExpiresAt.Before(time.Now()) {
			continue
		}

		// 提取访问密钥
		accessKey := fullKey[len(r.prefix)+1:]

		keyInfo := map[string]interface{}{
			"access_key":  accessKey,
			"cache_key":   info.CacheKey,
			"project_key": info.ProjectKey,
			"created_at":  info.CreatedAt,
			"expires_at":  info.ExpiresAt,
			"is_active":   info.IsActive,
		}

		keys = append(keys, keyInfo)
	}

	if err := iter.Err(); err != nil {
		r.logger.WithError(err).WithField("project_key", projectKey).Error("Failed to scan active access keys")
		return nil, fmt.Errorf("failed to scan active access keys: %w", err)
	}

	return keys, nil
}

// RevokeAccessKey 撤销访问密钥
func (r *RedisAccessKeyRepository) RevokeAccessKey(ctx context.Context, accessKey string) error {
	fullKey := r.buildKey(accessKey)

	// 获取当前信息
	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("access key not found: %s", accessKey)
		}
		return fmt.Errorf("failed to get access key: %w", err)
	}

	var info AccessKeyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return fmt.Errorf("failed to unmarshal access key info: %w", err)
	}

	// 更新为非活跃状态
	info.IsActive = false
	updatedData, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal updated access key info: %w", err)
	}

	// 计算剩余TTL
	ttl := info.ExpiresAt.Sub(time.Now())
	if ttl <= 0 {
		// 已过期，直接删除
		return r.DeleteAccessKey(ctx, accessKey)
	}

	// 更新密钥状态
	if err := r.client.Set(ctx, fullKey, updatedData, ttl).Err(); err != nil {
		r.logger.WithError(err).WithField("access_key", accessKey[:8]+"...").Error("Failed to revoke access key")
		return fmt.Errorf("failed to revoke access key: %w", err)
	}

	r.logger.WithField("access_key", accessKey[:8]+"...").Info("Access key revoked successfully")
	return nil
}

// RefreshAccessKey 刷新访问密钥
func (r *RedisAccessKeyRepository) RefreshAccessKey(ctx context.Context, accessKey string) (string, error) {
	fullKey := r.buildKey(accessKey)

	// 获取现有信息
	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("access key not found: %s", accessKey)
		}
		return "", fmt.Errorf("failed to get access key: %w", err)
	}

	var info AccessKeyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return "", fmt.Errorf("failed to unmarshal access key info: %w", err)
	}

	if !info.IsActive {
		return "", fmt.Errorf("cannot refresh inactive access key: %s", accessKey)
	}

	// 生成新的访问密钥
	serverKey := r.extractServerKey(info.CacheKey)
	newAccessKey := utils.GenerateAccessKey(serverKey, info.ProjectKey)

	// 保存新密钥
	if err := r.SaveAccessKey(ctx, info.CacheKey, newAccessKey); err != nil {
		return "", fmt.Errorf("failed to save new access key: %w", err)
	}

	// 撤销旧密钥
	if err := r.RevokeAccessKey(ctx, accessKey); err != nil {
		r.logger.WithError(err).WithField("old_access_key", accessKey[:8]+"...").Warn("Failed to revoke old access key")
	}

	r.logger.WithFields(map[string]interface{}{
		"old_access_key": accessKey[:8] + "...",
		"new_access_key": newAccessKey[:8] + "...",
		"project_key":    info.ProjectKey,
	}).Info("Access key refreshed successfully")

	return newAccessKey, nil
}

// extractServerKey 从缓存键中提取服务器键
func (r *RedisAccessKeyRepository) extractServerKey(cacheKey string) string {
	parts := utils.SplitString(cacheKey, ":")
	if len(parts) >= 1 {
		return parts[0]
	}
	return "default"
}

// GetAccessKeyInfo 获取访问密钥详细信息
func (r *RedisAccessKeyRepository) GetAccessKeyInfo(ctx context.Context, accessKey string) (*AccessKeyInfo, error) {
	fullKey := r.buildKey(accessKey)

	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("access key not found: %s", accessKey)
		}
		return nil, fmt.Errorf("failed to get access key: %w", err)
	}

	var info AccessKeyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal access key info: %w", err)
	}

	return &info, nil
}