package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheManager 缓存管理器
type CacheManager struct {
	client      *redis.Client
	enabled     bool
	redisConn   bool                  // Redis连接状态
	memoryCache map[string]*CacheItem // 内存缓存
	mu          sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	ExpireAt time.Time   `json:"expire_at"`
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(redisAddr string, redisPassword string, redisDB int) *CacheManager {
	cm := &CacheManager{
		enabled:     false, // 默认禁用，直到连接成功
		redisConn:   false,
		memoryCache: make(map[string]*CacheItem),
	}

	// 如果没有提供Redis地址，使用纯内存缓存
	if redisAddr == "" {
		log.Println("⚠️  Redis未配置，启用纯内存缓存模式")
		return cm
	}

	log.Printf("🔄 尝试连接Redis服务器: %s", redisAddr)
	client := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		MaxRetries:   2,
	})

	cm.client = client

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("❌ Redis连接失败，启用内存缓存降级模式: %v", err)
		log.Printf("💡 提示：Redis服务器未运行或配置错误，系统将使用内存缓存")
		log.Printf("🔧 如需启用Redis缓存，请检查Redis服务器状态和配置")
		return cm
	}

	// Redis连接成功
	cm.enabled = true
	cm.redisConn = true
	log.Printf("✅ Redis缓存已启用: %s", redisAddr)

	// 启动自动重连检查
	go cm.startHealthCheck()

	return cm
}

// Set 设置缓存
func (c *CacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果Redis可用，优先使用Redis
	if c.enabled && c.redisConn {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("缓存数据序列化失败: %w", err)
		}

		err = c.client.Set(ctx, key, data, expiration).Err()
		if err != nil {
			log.Printf("⚠️  Redis缓存设置失败，降级到内存缓存: %v", err)
			c.redisConn = false
			// 继续使用内存缓存
		} else {
			return nil
		}
	}

	// 使用内存缓存
	if expiration > 0 {
		c.memoryCache[key] = &CacheItem{
			Key:      key,
			Value:    value,
			ExpireAt: time.Now().Add(expiration),
		}
	} else {
		c.memoryCache[key] = &CacheItem{
			Key:      key,
			Value:    value,
			ExpireAt: time.Time{}, // 永不过期
		}
	}

	return nil
}

// Get 获取缓存
func (c *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果Redis可用，优先从Redis获取
	if c.enabled && c.redisConn {
		data, err := c.client.Get(ctx, key).Result()
		if err == nil {
			return json.Unmarshal([]byte(data), dest)
		} else if err != redis.Nil {
			log.Printf("⚠️  Redis缓存获取失败，降级到内存缓存: %v", err)
			c.redisConn = false
			// 继续从内存缓存获取
		}
	}

	// 从内存缓存获取
	if item, exists := c.memoryCache[key]; exists {
		// 检查是否过期
		if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
			delete(c.memoryCache, key)
			return fmt.Errorf("缓存已过期")
		}

		// 使用JSON序列化来复制值
		data, err := json.Marshal(item.Value)
		if err != nil {
			return fmt.Errorf("内存缓存数据序列化失败: %w", err)
		}

		err = json.Unmarshal(data, dest)
		if err != nil {
			return fmt.Errorf("内存缓存数据反序列化失败: %w", err)
		}

		return nil
	}

	return fmt.Errorf("缓存未命中")
}

// Delete 删除缓存
func (c *CacheManager) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return nil
	}

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("缓存删除失败: %v", err)
		return err
	}

	return nil
}

// Exists 检查缓存是否存在
func (c *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.enabled {
		return false, nil
	}

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("缓存检查失败: %v", err)
		return false, err
	}

	return count > 0, nil
}

// ClearPattern 清除匹配模式的缓存
func (c *CacheManager) ClearPattern(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return nil
	}

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("缓存模式匹配失败: %v", err)
		return err
	}

	if len(keys) > 0 {
		err = c.client.Del(ctx, keys...).Err()
		if err != nil {
			log.Printf("批量缓存删除失败: %v", err)
			return err
		}
		log.Printf("清除缓存模式 %s，删除 %d 个键", pattern, len(keys))
	}

	return nil
}

// GetStats 获取缓存统计信息
func (c *CacheManager) GetStats(ctx context.Context) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled": c.enabled,
		"mode":    "memory",
	}

	// 统计内存缓存信息
	memoryCount := len(c.memoryCache)
	expiredCount := 0
	now := time.Now()
	for _, item := range c.memoryCache {
		if !item.ExpireAt.IsZero() && now.After(item.ExpireAt) {
			expiredCount++
		}
	}

	stats["memory_cache"] = map[string]interface{}{
		"total_items":   memoryCount,
		"expired_items": expiredCount,
	}

	if c.enabled {
		stats["mode"] = "redis"
		stats["redis_connected"] = c.redisConn

		if c.redisConn && c.client != nil {
			info, err := c.client.Info(ctx).Result()
			if err != nil {
				log.Printf("获取Redis信息失败: %v", err)
				stats["redis_error"] = err.Error()
			} else {
				stats["redis_info"] = info
			}
		}
	}

	return stats, nil
}

// startHealthCheck 启动健康检查
func (c *CacheManager) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.checkRedisHealth()
		}
	}
}

// checkRedisHealth 检查Redis健康状态
func (c *CacheManager) checkRedisHealth() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled || c.client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		if c.redisConn {
			log.Printf("⚠️  Redis连接断开，启用内存缓存降级模式: %v", err)
			log.Printf("💡 系统将继续运行，但性能可能降低")
		}
		c.redisConn = false
		return
	}

	if !c.redisConn {
		log.Printf("✅ Redis连接恢复，切换回Redis缓存模式")
		c.redisConn = true
	}
}

// cleanupExpiredCache 清理过期缓存
func (c *CacheManager) cleanupExpiredCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.memoryCache {
		if !item.ExpireAt.IsZero() && now.After(item.ExpireAt) {
			delete(c.memoryCache, key)
		}
	}
}

// Close 关闭缓存连接
func (c *CacheManager) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.enabled && c.client != nil {
		return c.client.Close()
	}
	return nil
}

// 便捷方法：缓存常用数据
func (c *CacheManager) SetServersList(ctx context.Context, projectKey string, servers []ServerStatus) error {
	key := fmt.Sprintf("servers:list:%s", projectKey)
	return c.Set(ctx, key, servers, 30*time.Second) // 30秒缓存
}

func (c *CacheManager) GetServersList(ctx context.Context, projectKey string) ([]ServerStatus, error) {
	key := fmt.Sprintf("servers:list:%s", projectKey)
	var servers []ServerStatus
	err := c.Get(ctx, key, &servers)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func (c *CacheManager) SetServerCount(ctx context.Context, projectKey string, count int) error {
	key := fmt.Sprintf("servers:count:%s", projectKey)
	return c.Set(ctx, key, count, 30*time.Second)
}

func (c *CacheManager) GetServerCount(ctx context.Context, projectKey string) (int, error) {
	key := fmt.Sprintf("servers:count:%s", projectKey)
	var countStr string
	err := c.Get(ctx, key, &countStr)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(countStr)
}

func (c *CacheManager) SetUUIDStats(ctx context.Context, stats map[string]interface{}) error {
	return c.Set(ctx, "uuid:stats", stats, 60*time.Second) // 1分钟缓存
}

func (c *CacheManager) GetUUIDStats(ctx context.Context) (map[string]interface{}, error) {
	var stats map[string]interface{}
	err := c.Get(ctx, "uuid:stats", &stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (c *CacheManager) InvalidateServerCache(ctx context.Context, projectKey string) error {
	pattern := fmt.Sprintf("servers:*:%s", projectKey)
	return c.ClearPattern(ctx, pattern)
}

// 全局缓存管理器实例
var cacheManager *CacheManager
