package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// RedisCacheRepository Redis缓存仓库实现
type RedisCacheRepository struct {
	client *redis.Client
	logger logger.Logger
	prefix string
}

// NewRedisCacheRepository 创建Redis缓存仓库
func NewRedisCacheRepository(client *redis.Client, logger logger.Logger, prefix string) repository.CacheRepository {
	if prefix == "" {
		prefix = "serverstatus"
	}
	return &RedisCacheRepository{
		client: client,
		logger: logger,
		prefix: prefix,
	}
}

// buildKey 构建完整的键名
func (r *RedisCacheRepository) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

// Set 设置缓存
func (r *RedisCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := r.buildKey(key)

	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to marshal cache value")
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	// 设置缓存
	if err := r.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to set cache")
		return fmt.Errorf("failed to set cache: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"key": key,
		"ttl": ttl,
	}).Debug("Cache set successfully")

	return nil
}

// Get 获取缓存
func (r *RedisCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := r.buildKey(key)

	// 获取缓存值
	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache key not found: %s", key)
		}
		r.logger.WithError(err).WithField("key", key).Error("Failed to get cache")
		return fmt.Errorf("failed to get cache: %w", err)
	}

	// 反序列化值
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal cache value")
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (r *RedisCacheRepository) Delete(ctx context.Context, key string) error {
	fullKey := r.buildKey(key)

	if err := r.client.Del(ctx, fullKey).Err(); err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to delete cache")
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	r.logger.WithField("key", key).Debug("Cache deleted successfully")
	return nil
}

// Exists 检查缓存是否存在
func (r *RedisCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.buildKey(key)

	result, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to check cache existence")
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return result > 0, nil
}

// SetMultiple 批量设置缓存
func (r *RedisCacheRepository) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	// 使用Pipeline提高性能
	pipe := r.client.Pipeline()

	for key, value := range items {
		fullKey := r.buildKey(key)

		data, err := json.Marshal(value)
		if err != nil {
			r.logger.WithError(err).WithField("key", key).Error("Failed to marshal cache value")
			continue
		}

		pipe.Set(ctx, fullKey, data, ttl)
	}

	// 执行Pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.WithError(err).Error("Failed to execute cache pipeline")
		return fmt.Errorf("failed to execute cache pipeline: %w", err)
	}

	r.logger.WithField("count", len(items)).Debug("Multiple cache items set successfully")
	return nil
}

// GetMultiple 批量获取缓存
func (r *RedisCacheRepository) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 构建完整的键名
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	// 批量获取
	results, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get multiple cache items")
		return nil, fmt.Errorf("failed to get multiple cache items: %w", err)
	}

	// 处理结果
	response := make(map[string]interface{})
	for i, key := range keys {
		if results[i] != nil {
			data, ok := results[i].(string)
			if !ok {
				r.logger.WithField("key", key).Warn("Cache value is not a string")
				continue
			}

			var dest interface{}
			if err := json.Unmarshal([]byte(data), &dest); err != nil {
				r.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal cache value")
				continue
			}

			response[key] = dest
		}
	}

	r.logger.WithFields(map[string]interface{}{
		"requested": len(keys),
		"found":     len(response),
	}).Debug("Multiple cache items retrieved")

	return response, nil
}

// DeleteMultiple 批量删除缓存
func (r *RedisCacheRepository) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// 构建完整的键名
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	if err := r.client.Del(ctx, fullKeys...).Err(); err != nil {
		r.logger.WithError(err).Error("Failed to delete multiple cache items")
		return fmt.Errorf("failed to delete multiple cache items: %w", err)
	}

	r.logger.WithField("count", len(keys)).Debug("Multiple cache items deleted successfully")
	return nil
}

// ClearPattern 清除匹配模式的缓存
func (r *RedisCacheRepository) ClearPattern(ctx context.Context, pattern string) error {
	fullPattern := r.buildKey(pattern)

	// 使用SCAN而不是KEYS以避免阻塞
	iter := r.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		r.logger.WithError(err).WithField("pattern", pattern).Error("Failed to scan cache keys")
		return fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			r.logger.WithError(err).WithField("pattern", pattern).Error("Failed to delete cache pattern")
			return fmt.Errorf("failed to delete cache pattern: %w", err)
		}

		r.logger.WithFields(map[string]interface{}{
			"pattern": pattern,
			"deleted": len(keys),
		}).Debug("Cache pattern cleared successfully")
	}

	return nil
}

// Keys 获取匹配模式的键
func (r *RedisCacheRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	fullPattern := r.buildKey(pattern)

	iter := r.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		fullKey := iter.Val()
		// 移除前缀，返回原始键名
		if len(fullKey) > len(r.prefix)+1 {
			key := fullKey[len(r.prefix)+1:]
			keys = append(keys, key)
		}
	}

	if err := iter.Err(); err != nil {
		r.logger.WithError(err).WithField("pattern", pattern).Error("Failed to scan cache keys")
		return nil, fmt.Errorf("failed to scan cache keys: %w", err)
	}

	return keys, nil
}

// GetStats 获取缓存统计信息
func (r *RedisCacheRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取Redis信息
	info, err := r.client.Info(ctx, "memory", "keyspace", "stats").Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get Redis info")
		return stats, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// 解析内存信息
	memoryInfo := r.parseRedisInfo(info, "memory")
	if usedMemory, ok := memoryInfo["used_memory"]; ok {
		if memory, err := strconv.ParseInt(usedMemory, 10, 64); err == nil {
			stats["used_memory_bytes"] = memory
			stats["used_memory_mb"] = float64(memory) / (1024 * 1024)
		}
	}

	// 解析键空间信息
	keyspaceInfo := r.parseRedisInfo(info, "db")
	for db, info := range keyspaceInfo {
		stats[fmt.Sprintf("%s_keys", db)] = info
	}

	// 获取连接池状态
	poolStats := r.client.PoolStats()
	stats["connections"] = map[string]interface{}{
		"hits":        poolStats.Hits,
		"misses":      poolStats.Misses,
		"timeouts":    poolStats.Timeouts,
		"total_conns": poolStats.TotalConns,
		"idle_conns":  poolStats.IdleConns,
		"stale_conns": poolStats.StaleConns,
	}

	// 统计当前前缀下的键数量
	pattern := r.buildKey("*")
	_, err = r.client.ClientList(ctx).Result()
	if err == nil {
		stats["prefix_pattern"] = pattern
		// 这里可以添加更多自定义统计
	}

	stats["type"] = "redis"
	stats["prefix"] = r.prefix

	return stats, nil
}

// parseRedisInfo 解析Redis INFO命令的输出
func (r *RedisCacheRepository) parseRedisInfo(info, section string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")

	inSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			if section != "" && line[0] == '#' {
				inSection = false
			}
			continue
		}

		if section != "" {
			if line == section {
				inSection = true
				continue
			}
			if !inSection {
				continue
			}
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

// IsConnected 检查连接状态
func (r *RedisCacheRepository) IsConnected() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Ping(ctx).Err() == nil
}

// GetType 获取缓存类型
func (r *RedisCacheRepository) GetType() string {
	return "redis"
}

// HealthChecker 健康检查器
func (r *RedisCacheRepository) HealthChecker() repository.HealthChecker {
	return &RedisHealthChecker{
		client: r.client,
		logger: r.logger,
	}
}

// RedisHealthChecker Redis健康检查器
type RedisHealthChecker struct {
	client *redis.Client
	logger logger.Logger
}

// CheckHealth 检查健康状态
func (h *RedisHealthChecker) CheckHealth(ctx context.Context) error {
	// 测试基本连接
	if err := h.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// 测试读写操作
	testKey := "health_check_test"
	testValue := map[string]interface{}{"test": true, "timestamp": time.Now()}

	// 写入测试数据
	if err := h.client.Set(ctx, testKey, testValue, 10*time.Second).Err(); err != nil {
		return fmt.Errorf("redis write test failed: %w", err)
	}

	// 读取测试数据
	result, err := h.client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("redis read test failed: %w", err)
	}

	// 清理测试数据
	h.client.Del(ctx, testKey)

	if result == "" {
		return fmt.Errorf("redis data integrity test failed")
	}

	return nil
}

// GetMetrics 获取Redis指标
func (h *RedisHealthChecker) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	info, err := h.client.Info(ctx, "server", "memory", "clients", "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis info: %w", err)
	}

	metrics := make(map[string]interface{})

	// 解析服务器信息
	serverInfo := h.parseInfo(info, "server")
	for key, value := range serverInfo {
		metrics[fmt.Sprintf("server_%s", key)] = value
	}

	// 解析内存信息
	memoryInfo := h.parseInfo(info, "memory")
	for key, value := range memoryInfo {
		metrics[fmt.Sprintf("memory_%s", key)] = value
	}

	// 解析客户端信息
	clientsInfo := h.parseInfo(info, "clients")
	for key, value := range clientsInfo {
		metrics[fmt.Sprintf("clients_%s", key)] = value
	}

	// 解析统计信息
	statsInfo := h.parseInfo(info, "stats")
	for key, value := range statsInfo {
		metrics[fmt.Sprintf("stats_%s", key)] = value
	}

	// 添加连接池状态
	poolStats := h.client.PoolStats()
	metrics["pool_hits"] = poolStats.Hits
	metrics["pool_misses"] = poolStats.Misses
	metrics["pool_timeouts"] = poolStats.Timeouts
	metrics["pool_total_conns"] = poolStats.TotalConns
	metrics["pool_idle_conns"] = poolStats.IdleConns
	metrics["pool_stale_conns"] = poolStats.StaleConns

	return metrics, nil
}

// parseInfo 解析INFO命令输出
func (h *RedisHealthChecker) parseInfo(info, section string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")

	inSection := section == ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			if section != "" && line[0] == '#' {
				if strings.Contains(line, section) {
					inSection = true
				} else if inSection {
					inSection = false
				}
			}
			continue
		}

		if section != "" && !inSection {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}
