package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// SQLiteCacheRepository SQLite缓存仓库实现
type SQLiteCacheRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewSQLiteCacheRepository 创建SQLite缓存仓库
func NewSQLiteCacheRepository(db *sql.DB, logger logger.Logger) repository.CacheRepository {
	return &SQLiteCacheRepository{
		db:     db,
		logger: logger,
	}
}

// Set 设置缓存
func (r *SQLiteCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to marshal cache value")
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	// 计算过期时间
	expiresAt := time.Now().Add(ttl)

	// 插入或更新缓存
	query := `
		INSERT OR REPLACE INTO cache (key, value, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query, key, string(data), expiresAt, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to set cache")
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Get 获取缓存
func (r *SQLiteCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	query := `
		SELECT value, expires_at
		FROM cache
		WHERE key = ? AND (expires_at IS NULL OR expires_at > ?)
	`

	var value string
	var expiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, key, time.Now()).Scan(&value, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("cache key not found: %s", key)
		}
		r.logger.WithError(err).WithField("key", key).Error("Failed to get cache")
		return fmt.Errorf("failed to get cache: %w", err)
	}

	// 反序列化值
	if err := json.Unmarshal([]byte(value), dest); err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal cache value")
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (r *SQLiteCacheRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM cache WHERE key = ?`

	_, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to delete cache")
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// Exists 检查缓存是否存在
func (r *SQLiteCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM cache
		WHERE key = ? AND (expires_at IS NULL OR expires_at > ?)
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, key, time.Now()).Scan(&count)
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Failed to check cache existence")
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return count > 0, nil
}

// SetMultiple 批量设置缓存
func (r *SQLiteCacheRepository) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	// 开始事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT OR REPLACE INTO cache (key, value, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`

	expiresAt := time.Now().Add(ttl)
	createdAt := time.Now()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			r.logger.WithError(err).WithField("key", key).Error("Failed to marshal cache value")
			continue
		}

		_, err = tx.ExecContext(ctx, query, key, string(data), expiresAt, createdAt)
		if err != nil {
			r.logger.WithError(err).WithField("key", key).Error("Failed to set cache in transaction")
			continue
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetMultiple 批量获取缓存
func (r *SQLiteCacheRepository) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 构建查询
	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys)+1)
	args[0] = time.Now() // 过期时间参数

	for i, key := range keys {
		placeholders[i] = "?"
		args[i+1] = key
	}

	query := fmt.Sprintf(`
		SELECT key, value
		FROM cache
		WHERE key IN (%s) AND (expires_at IS NULL OR expires_at > ?)
	`, fmt.Sprintf("%s", placeholders))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query multiple cache items")
		return nil, fmt.Errorf("failed to query multiple cache items: %w", err)
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			r.logger.WithError(err).Error("Failed to scan cache row")
			continue
		}

		var dest interface{}
		if err := json.Unmarshal([]byte(value), &dest); err != nil {
			r.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal cache value")
			continue
		}

		result[key] = dest
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating cache rows")
		return nil, fmt.Errorf("error iterating cache rows: %w", err)
	}

	return result, nil
}

// DeleteMultiple 批量删除缓存
func (r *SQLiteCacheRepository) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// 构建查询
	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))

	for i, key := range keys {
		placeholders[i] = "?"
		args[i] = key
	}

	query := fmt.Sprintf(`DELETE FROM cache WHERE key IN (%s)`, fmt.Sprintf("%s", placeholders))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete multiple cache items")
		return fmt.Errorf("failed to delete multiple cache items: %w", err)
	}

	return nil
}

// ClearPattern 清除匹配模式的缓存
func (r *SQLiteCacheRepository) ClearPattern(ctx context.Context, pattern string) error {
	// SQLite使用LIKE进行模式匹配
	query := `DELETE FROM cache WHERE key LIKE ?`

	_, err := r.db.ExecContext(ctx, query, pattern)
	if err != nil {
		r.logger.WithError(err).WithField("pattern", pattern).Error("Failed to clear cache pattern")
		return fmt.Errorf("failed to clear cache pattern: %w", err)
	}

	return nil
}

// Keys 获取匹配模式的键
func (r *SQLiteCacheRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	query := `
		SELECT key
		FROM cache
		WHERE key LIKE ? AND (expires_at IS NULL OR expires_at > ?)
	`

	rows, err := r.db.QueryContext(ctx, query, pattern, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("pattern", pattern).Error("Failed to query cache keys")
		return nil, fmt.Errorf("failed to query cache keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			r.logger.WithError(err).Error("Failed to scan cache key")
			continue
		}
		keys = append(keys, key)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating cache keys")
		return nil, fmt.Errorf("error iterating cache keys: %w", err)
	}

	return keys, nil
}

// GetStats 获取缓存统计信息
func (r *SQLiteCacheRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总缓存数量
	var totalCount int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM cache`).Scan(&totalCount)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get total cache count")
	} else {
		stats["total_keys"] = totalCount
	}

	// 有效缓存数量
	var activeCount int64
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM cache WHERE expires_at IS NULL OR expires_at > ?`, time.Now()).Scan(&activeCount)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get active cache count")
	} else {
		stats["active_keys"] = activeCount
	}

	// 过期缓存数量
	stats["expired_keys"] = totalCount - activeCount

	// 数据库大小
	var size int64
	err = r.db.QueryRowContext(ctx, `SELECT SUM(LENGTH(value)) FROM cache`).Scan(&size)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get cache size")
	} else {
		stats["size_bytes"] = size
		stats["size_mb"] = float64(size) / (1024 * 1024)
	}

	return stats, nil
}

// IsConnected 检查连接状态
func (r *SQLiteCacheRepository) IsConnected() bool {
	return r.db.Ping() == nil
}

// GetType 获取缓存类型
func (r *SQLiteCacheRepository) GetType() string {
	return "sqlite"
}