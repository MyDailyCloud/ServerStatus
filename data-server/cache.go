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
	client  *redis.Client
	enabled bool
	mu      sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	ExpireAt time.Time   `json:"expire_at"`
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(redisAddr string, redisPassword string, redisDB int) *CacheManager {
	// 如果没有提供Redis地址，创建一个内存缓存管理器
	if redisAddr == "" {
		log.Println("Redis未配置，使用内存缓存")
		return &CacheManager{
			client:  nil,
			enabled: false,
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis连接失败，使用内存缓存: %v", err)
		return &CacheManager{
			client:  nil,
			enabled: false,
		}
	}

	log.Printf("Redis缓存已启用: %s", redisAddr)
	return &CacheManager{
		client:  client,
		enabled: true,
	}
}

// Set 设置缓存
func (c *CacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return nil // 如果缓存未启用，直接返回
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("缓存数据序列化失败: %w", err)
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		log.Printf("缓存设置失败: %v", err)
		return err
	}

	return nil
}

// Get 获取缓存
func (c *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.enabled {
		return fmt.Errorf("缓存未启用") // 如果缓存未启用，返回错误
	}

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("缓存未命中")
		}
		log.Printf("缓存获取失败: %v", err)
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("缓存数据反序列化失败: %w", err)
	}

	return nil
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
	}

	if !c.enabled {
		return stats, nil
	}

	info, err := c.client.Info(ctx).Result()
	if err != nil {
		log.Printf("获取Redis信息失败: %v", err)
		return stats, nil
	}

	// 解析Redis信息获取简单的统计
	stats["redis_info"] = info
	stats["connected_clients"] = "N/A" // 可以进一步解析

	return stats, nil
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