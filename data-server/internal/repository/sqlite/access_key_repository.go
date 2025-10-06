package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
	"github.com/kanshan/ServerStatus/data-server/pkg/utils"
)

// SQLiteAccessKeyRepository SQLite访问密钥仓库实现
type SQLiteAccessKeyRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewSQLiteAccessKeyRepository 创建SQLite访问密钥仓库
func NewSQLiteAccessKeyRepository(db *sql.DB, logger logger.Logger) repository.AccessKeyRepository {
	return &SQLiteAccessKeyRepository{
		db:     db,
		logger: logger,
	}
}

// SaveAccessKey 保存访问密钥
func (r *SQLiteAccessKeyRepository) SaveAccessKey(ctx context.Context, cacheKey, accessKey string) error {
	query := `
		INSERT OR REPLACE INTO access_keys (access_key, cache_key, project_key, created_at, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	// 解析cacheKey获取projectKey
	projectKey := r.extractProjectKey(cacheKey)
	expiresAt := time.Now().Add(24 * time.Hour) // 默认24小时过期

	_, err := r.db.ExecContext(ctx, query, accessKey, cacheKey, projectKey, time.Now(), expiresAt, true)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"access_key": accessKey[:8] + "...",
			"cache_key":  cacheKey,
		}).Error("Failed to save access key")
		return fmt.Errorf("failed to save access key: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"access_key": accessKey[:8] + "...",
		"project_key": projectKey,
	}).Info("Access key saved successfully")

	return nil
}

// GetAccessKey 获取访问密钥
func (r *SQLiteAccessKeyRepository) GetAccessKey(ctx context.Context, accessKey string) (string, error) {
	query := `
		SELECT cache_key, project_key, expires_at, is_active
		FROM access_keys
		WHERE access_key = ?
	`

	var cacheKey, projectKey string
	var expiresAt sql.NullTime
	var isActive bool

	err := r.db.QueryRowContext(ctx, query, accessKey).Scan(&cacheKey, &projectKey, &expiresAt, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("access key not found: %s", accessKey)
		}
		r.logger.WithError(err).WithField("access_key", accessKey[:8]+"...").Error("Failed to get access key")
		return "", fmt.Errorf("failed to get access key: %w", err)
	}

	// 检查是否有效
	if !isActive {
		return "", fmt.Errorf("access key is inactive: %s", accessKey)
	}

	// 检查是否过期
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return "", fmt.Errorf("access key expired: %s", accessKey)
	}

	return projectKey, nil
}

// DeleteAccessKey 删除访问密钥
func (r *SQLiteAccessKeyRepository) DeleteAccessKey(ctx context.Context, accessKey string) error {
	query := `DELETE FROM access_keys WHERE access_key = ?`

	_, err := r.db.ExecContext(ctx, query, accessKey)
	if err != nil {
		r.logger.WithError(err).WithField("access_key", accessKey[:8]+"...").Error("Failed to delete access key")
		return fmt.Errorf("failed to delete access key: %w", err)
	}

	r.logger.WithField("access_key", accessKey[:8]+"...").Info("Access key deleted successfully")
	return nil
}

// GenerateAccessKey 生成访问密钥
func (r *SQLiteAccessKeyRepository) GenerateAccessKey(ctx context.Context, serverKey, projectKey string) (string, error) {
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
func (r *SQLiteAccessKeyRepository) ValidateAccessKey(ctx context.Context, accessKey string) (string, error) {
	projectKey, err := r.GetAccessKey(ctx, accessKey)
	if err != nil {
		return "", err
	}

	return projectKey, nil
}

// CleanupExpiredKeys 清理过期密钥
func (r *SQLiteAccessKeyRepository) CleanupExpiredKeys(ctx context.Context, ttl time.Duration) error {
	query := `DELETE FROM access_keys WHERE expires_at < datetime('now', '-' || ? || ' seconds')`

	ttlSeconds := int(ttl.Seconds())
	result, err := r.db.ExecContext(ctx, query, ttlSeconds)
	if err != nil {
		r.logger.WithError(err).Error("Failed to cleanup expired access keys")
		return fmt.Errorf("failed to cleanup expired access keys: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.logger.WithField("rows_affected", rowsAffected).Info("Expired access keys cleaned up")
	}

	return nil
}

// extractProjectKey 从缓存键中提取项目键
func (r *SQLiteAccessKeyRepository) extractProjectKey(cacheKey string) string {
	// 简单的解析逻辑，实际可能需要更复杂的处理
	parts := utils.SplitString(cacheKey, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "default"
}

// GetActiveKeys 获取活跃的访问密钥
func (r *SQLiteAccessKeyRepository) GetActiveKeys(ctx context.Context, projectKey string) ([]map[string]interface{}, error) {
	query := `
		SELECT access_key, cache_key, project_key, created_at, expires_at, is_active
		FROM access_keys
		WHERE (project_key = ? OR ? = '') AND is_active = 1
		  AND (expires_at IS NULL OR expires_at > ?)
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectKey, projectKey, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("project_key", projectKey).Error("Failed to get active access keys")
		return nil, fmt.Errorf("failed to get active access keys: %w", err)
	}
	defer rows.Close()

	var keys []map[string]interface{}
	for rows.Next() {
		var accessKey, cacheKey, pKey string
		var createdAt, expiresAt sql.NullTime
		var isActive bool

		err := rows.Scan(&accessKey, &cacheKey, &pKey, &createdAt, &expiresAt, &isActive)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan access key row")
			continue
		}

		keyInfo := map[string]interface{}{
			"access_key": accessKey,
			"cache_key":  cacheKey,
			"project_key": pKey,
			"is_active":  isActive,
		}

		if createdAt.Valid {
			keyInfo["created_at"] = createdAt.Time
		}

		if expiresAt.Valid {
			keyInfo["expires_at"] = expiresAt.Time
		}

		keys = append(keys, keyInfo)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating access key rows")
		return nil, fmt.Errorf("error iterating access key rows: %w", err)
	}

	return keys, nil
}

// RevokeAccessKey 撤销访问密钥
func (r *SQLiteAccessKeyRepository) RevokeAccessKey(ctx context.Context, accessKey string) error {
	query := `UPDATE access_keys SET is_active = 0 WHERE access_key = ?`

	_, err := r.db.ExecContext(ctx, query, accessKey)
	if err != nil {
		r.logger.WithError(err).WithField("access_key", accessKey[:8]+"...").Error("Failed to revoke access key")
		return fmt.Errorf("failed to revoke access key: %w", err)
	}

	r.logger.WithField("access_key", accessKey[:8]+"...").Info("Access key revoked successfully")
	return nil
}

// RefreshAccessKey 刷新访问密钥
func (r *SQLiteAccessKeyRepository) RefreshAccessKey(ctx context.Context, accessKey string) (string, error) {
	// 获取现有信息
	query := `
		SELECT cache_key, project_key, is_active
		FROM access_keys
		WHERE access_key = ?
	`

	var cacheKey, projectKey string
	var isActive bool

	err := r.db.QueryRowContext(ctx, query, accessKey).Scan(&cacheKey, &projectKey, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("access key not found: %s", accessKey)
		}
		return "", fmt.Errorf("failed to get access key: %w", err)
	}

	if !isActive {
		return "", fmt.Errorf("cannot refresh inactive access key: %s", accessKey)
	}

	// 生成新的访问密钥
	serverKey := r.extractServerKey(cacheKey)
	newAccessKey := utils.GenerateAccessKey(serverKey, projectKey)

	// 保存新密钥并使旧密钥失效
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 使旧密钥失效
	_, err = tx.ExecContext(ctx, `UPDATE access_keys SET is_active = 0 WHERE access_key = ?`, accessKey)
	if err != nil {
		return "", fmt.Errorf("failed to deactivate old access key: %w", err)
	}

	// 保存新密钥
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO access_keys (access_key, cache_key, project_key, created_at, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, newAccessKey, cacheKey, projectKey, time.Now(), expiresAt, true)
	if err != nil {
		return "", fmt.Errorf("failed to save new access key: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"old_access_key": accessKey[:8] + "...",
		"new_access_key": newAccessKey[:8] + "...",
		"project_key":    projectKey,
	}).Info("Access key refreshed successfully")

	return newAccessKey, nil
}

// extractServerKey 从缓存键中提取服务器键
func (r *SQLiteAccessKeyRepository) extractServerKey(cacheKey string) string {
	parts := utils.SplitString(cacheKey, ":")
	if len(parts) >= 1 {
		return parts[0]
	}
	return "default"
}