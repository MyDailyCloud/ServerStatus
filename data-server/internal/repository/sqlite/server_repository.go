package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
	"github.com/kanshan/ServerStatus/data-server/internal/repository"
	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// SQLiteServerRepository SQLite服务器数据仓库实现
type SQLiteServerRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewSQLiteServerRepository 创建SQLite服务器仓库
func NewSQLiteServerRepository(db *sql.DB, logger logger.Logger) repository.ServerRepository {
	return &SQLiteServerRepository{
		db:     db,
		logger: logger,
	}
}

// CreateServer 创建服务器记录
func (r *SQLiteServerRepository) CreateServer(ctx context.Context, server *models.ServerInfo) error {
	query := `
		INSERT INTO servers (
			session_id, hostname, project_key, os, arch, cpu_cores,
			memory_total, disk_total, uptime, boot_time, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		server.SessionID,
		server.Hostname,
		server.ProjectKey,
		server.OS,
		server.Arch,
		server.CPUCores,
		server.MemoryTotal,
		server.DiskTotal,
		server.Uptime,
		server.BootTime,
		now,
		now,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create server record")
		return fmt.Errorf("failed to create server: %w", err)
	}

	r.logger.WithField("session_id", server.SessionID).Info("Server record created successfully")
	return nil
}

// GetServer 获取服务器信息
func (r *SQLiteServerRepository) GetServer(ctx context.Context, sessionID string) (*models.ServerInfo, error) {
	query := `
		SELECT session_id, hostname, project_key, os, arch, cpu_cores,
			   memory_total, disk_total, uptime, boot_time, created_at, updated_at
		FROM servers
		WHERE session_id = ?
	`

	server := &models.ServerInfo{}
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&server.SessionID,
		&server.Hostname,
		&server.ProjectKey,
		&server.OS,
		&server.Arch,
		&server.CPUCores,
		&server.MemoryTotal,
		&server.DiskTotal,
		&server.Uptime,
		&server.BootTime,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server not found: %s", sessionID)
		}
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to get server")
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	server.CreatedAt = createdAt
	server.UpdatedAt = updatedAt

	// 获取最新状态信息
	status, err := r.getLatestServerStatus(ctx, sessionID)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Warn("Failed to get latest server status")
	} else {
		server.SystemInfo = status
	}

	return server, nil
}

// UpdateServer 更新服务器信息
func (r *SQLiteServerRepository) UpdateServer(ctx context.Context, server *models.ServerInfo) error {
	query := `
		UPDATE servers
		SET hostname = ?, project_key = ?, os = ?, arch = ?, cpu_cores = ?,
			memory_total = ?, disk_total = ?, uptime = ?, boot_time = ?, updated_at = ?
		WHERE session_id = ?
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		server.Hostname,
		server.ProjectKey,
		server.OS,
		server.Arch,
		server.CPUCores,
		server.MemoryTotal,
		server.DiskTotal,
		server.Uptime,
		server.BootTime,
		now,
		server.SessionID,
	)

	if err != nil {
		r.logger.WithError(err).WithField("session_id", server.SessionID).Error("Failed to update server")
		return fmt.Errorf("failed to update server: %w", err)
	}

	r.logger.WithField("session_id", server.SessionID).Info("Server record updated successfully")
	return nil
}

// DeleteServer 删除服务器记录
func (r *SQLiteServerRepository) DeleteServer(ctx context.Context, sessionID string) error {
	// 删除相关的历史数据
	_, err := r.db.ExecContext(ctx, "DELETE FROM server_history WHERE session_id = ?", sessionID)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to delete server history")
		return fmt.Errorf("failed to delete server history: %w", err)
	}

	// 删除服务器记录
	_, err = r.db.ExecContext(ctx, "DELETE FROM servers WHERE session_id = ?", sessionID)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to delete server")
		return fmt.Errorf("failed to delete server: %w", err)
	}

	r.logger.WithField("session_id", sessionID).Info("Server record deleted successfully")
	return nil
}

// GetAllServers 获取所有服务器
func (r *SQLiteServerRepository) GetAllServers(ctx context.Context, projectKey string, offset, limit int) ([]*models.ServerInfo, error) {
	query := `
		SELECT session_id, hostname, project_key, os, arch, cpu_cores,
			   memory_total, disk_total, uptime, boot_time, created_at, updated_at
		FROM servers
		WHERE project_key = ? OR ? = ''
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, projectKey, projectKey, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query servers")
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	var servers []*models.ServerInfo
	for rows.Next() {
		server := &models.ServerInfo{}
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&server.SessionID,
			&server.Hostname,
			&server.ProjectKey,
			&server.OS,
			&server.Arch,
			&server.CPUCores,
			&server.MemoryTotal,
			&server.DiskTotal,
			&server.Uptime,
			&server.BootTime,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan server row")
			continue
		}

		server.CreatedAt = createdAt
		server.UpdatedAt = updatedAt

		// 获取最新状态信息
		status, err := r.getLatestServerStatus(ctx, server.SessionID)
		if err != nil {
			r.logger.WithError(err).WithField("session_id", server.SessionID).Warn("Failed to get latest server status")
		} else {
			server.SystemInfo = status
		}

		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating server rows")
		return nil, fmt.Errorf("error iterating server rows: %w", err)
	}

	return servers, nil
}

// GetServersByHostname 根据主机名获取服务器
func (r *SQLiteServerRepository) GetServersByHostname(ctx context.Context, hostname string) ([]*models.ServerInfo, error) {
	query := `
		SELECT session_id, hostname, project_key, os, arch, cpu_cores,
			   memory_total, disk_total, uptime, boot_time, created_at, updated_at
		FROM servers
		WHERE hostname = ?
		ORDER BY updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, hostname)
	if err != nil {
		r.logger.WithError(err).WithField("hostname", hostname).Error("Failed to query servers by hostname")
		return nil, fmt.Errorf("failed to query servers by hostname: %w", err)
	}
	defer rows.Close()

	var servers []*models.ServerInfo
	for rows.Next() {
		server := &models.ServerInfo{}
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&server.SessionID,
			&server.Hostname,
			&server.ProjectKey,
			&server.OS,
			&server.Arch,
			&server.CPUCores,
			&server.MemoryTotal,
			&server.DiskTotal,
			&server.Uptime,
			&server.BootTime,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan server row")
			continue
		}

		server.CreatedAt = createdAt
		server.UpdatedAt = updatedAt

		// 获取最新状态信息
		status, err := r.getLatestServerStatus(ctx, server.SessionID)
		if err != nil {
			r.logger.WithError(err).WithField("session_id", server.SessionID).Warn("Failed to get latest server status")
		} else {
			server.SystemInfo = status
		}

		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating server rows")
		return nil, fmt.Errorf("error iterating server rows: %w", err)
	}

	return servers, nil
}

// GetServerCount 获取服务器数量
func (r *SQLiteServerRepository) GetServerCount(ctx context.Context, projectKey string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM servers
		WHERE project_key = ? OR ? = ''
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, projectKey, projectKey).Scan(&count)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count servers")
		return 0, fmt.Errorf("failed to count servers: %w", err)
	}

	return count, nil
}

// UpdateLastSeen 更新最后访问时间
func (r *SQLiteServerRepository) UpdateLastSeen(ctx context.Context, sessionID string) error {
	query := `UPDATE servers SET updated_at = ? WHERE session_id = ?`

	_, err := r.db.ExecContext(ctx, query, time.Now(), sessionID)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to update last seen")
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	return nil
}

// GetOnlineServers 获取在线服务器
func (r *SQLiteServerRepository) GetOnlineServers(ctx context.Context, projectKey string, timeout time.Duration) ([]*models.ServerInfo, error) {
	query := `
		SELECT session_id, hostname, project_key, os, arch, cpu_cores,
			   memory_total, disk_total, uptime, boot_time, created_at, updated_at
		FROM servers
		WHERE (project_key = ? OR ? = '')
		  AND updated_at > datetime('now', '-' || ? || ' seconds')
		ORDER BY updated_at DESC
	`

	timeoutSeconds := int(timeout.Seconds())
	rows, err := r.db.QueryContext(ctx, query, projectKey, projectKey, timeoutSeconds)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query online servers")
		return nil, fmt.Errorf("failed to query online servers: %w", err)
	}
	defer rows.Close()

	var servers []*models.ServerInfo
	for rows.Next() {
		server := &models.ServerInfo{}
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&server.SessionID,
			&server.Hostname,
			&server.ProjectKey,
			&server.OS,
			&server.Arch,
			&server.CPUCores,
			&server.MemoryTotal,
			&server.DiskTotal,
			&server.Uptime,
			&server.BootTime,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan server row")
			continue
		}

		server.CreatedAt = createdAt
		server.UpdatedAt = updatedAt
		server.IsOnline = true

		// 获取最新状态信息
		status, err := r.getLatestServerStatus(ctx, server.SessionID)
		if err != nil {
			r.logger.WithError(err).WithField("session_id", server.SessionID).Warn("Failed to get latest server status")
		} else {
			server.SystemInfo = status
		}

		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating server rows")
		return nil, fmt.Errorf("error iterating server rows: %w", err)
	}

	return servers, nil
}

// Ping 检查数据库连接
func (r *SQLiteServerRepository) Ping() error {
	return r.db.Ping()
}

// Close 关闭数据库连接
func (r *SQLiteServerRepository) Close() error {
	return r.db.Close()
}

// getLatestServerStatus 获取服务器最新状态
func (r *SQLiteServerRepository) getLatestServerStatus(ctx context.Context, sessionID string) (*models.SystemInfo, error) {
	query := `
		SELECT cpu_usage, memory_used, memory_available, disk_used, disk_available,
			   network_rx, network_tx, load_avg, process_count, timestamp
		FROM server_history
		WHERE session_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	status := &models.SystemInfo{}
	var timestamp time.Time

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&status.CPUUsage,
		&status.MemoryUsed,
		&status.MemoryAvailable,
		&status.DiskUsed,
		&status.DiskAvailable,
		&status.NetworkRx,
		&status.NetworkTx,
		&status.LoadAvg,
		&status.ProcessCount,
		&timestamp,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有历史数据
		}
		return nil, fmt.Errorf("failed to get latest server status: %w", err)
	}

	status.Timestamp = timestamp
	return status, nil
}
