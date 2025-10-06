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

// SQLiteHistoryRepository SQLite历史数据仓库实现
type SQLiteHistoryRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewSQLiteHistoryRepository 创建SQLite历史数据仓库
func NewSQLiteHistoryRepository(db *sql.DB, logger logger.Logger) repository.HistoryRepository {
	return &SQLiteHistoryRepository{
		db:     db,
		logger: logger,
	}
}

// SaveHistoryData 保存历史数据
func (r *SQLiteHistoryRepository) SaveHistoryData(ctx context.Context, data *models.SystemInfo) error {
	query := `
		INSERT INTO server_history (
			session_id, hostname, project_key, cpu_usage, memory_used, memory_available,
			disk_used, disk_available, network_rx, network_tx, load_avg, process_count, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		data.SessionID,
		data.Hostname,
		data.ProjectKey,
		data.CPUUsage,
		data.MemoryUsed,
		data.MemoryAvailable,
		data.DiskUsed,
		data.DiskAvailable,
		data.NetworkRx,
		data.NetworkTx,
		data.LoadAvg,
		data.ProcessCount,
		data.Timestamp,
	)

	if err != nil {
		r.logger.WithError(err).WithField("session_id", data.SessionID).Error("Failed to save history data")
		return fmt.Errorf("failed to save history data: %w", err)
	}

	return nil
}

// GetHostHistory 获取主机历史数据
func (r *SQLiteHistoryRepository) GetHostHistory(ctx context.Context, hostname, projectKey string, limit int) ([]*models.HistoryData, error) {
	query := `
		SELECT session_id, hostname, project_key, cpu_usage, memory_used, memory_available,
			   disk_used, disk_available, network_rx, network_tx, load_avg, process_count, timestamp
		FROM server_history
		WHERE hostname = ? AND (project_key = ? OR ? = '')
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, hostname, projectKey, projectKey, limit)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"hostname":    hostname,
			"project_key": projectKey,
		}).Error("Failed to query host history")
		return nil, fmt.Errorf("failed to query host history: %w", err)
	}
	defer rows.Close()

	var historyData []*models.HistoryData
	for rows.Next() {
		data := &models.HistoryData{}
		err := rows.Scan(
			&data.SessionID,
			&data.Hostname,
			&data.ProjectKey,
			&data.CPUUsage,
			&data.MemoryUsed,
			&data.MemoryAvailable,
			&data.DiskUsed,
			&data.DiskAvailable,
			&data.NetworkRx,
			&data.NetworkTx,
			&data.LoadAvg,
			&data.ProcessCount,
			&data.Timestamp,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan history row")
			continue
		}

		historyData = append(historyData, data)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating history rows")
		return nil, fmt.Errorf("error iterating history rows: %w", err)
	}

	return historyData, nil
}

// GetHistoryByTimeRange 获取时间范围内的历史数据
func (r *SQLiteHistoryRepository) GetHistoryByTimeRange(ctx context.Context, hostname, projectKey string, start, end time.Time) ([]*models.HistoryData, error) {
	query := `
		SELECT session_id, hostname, project_key, cpu_usage, memory_used, memory_available,
			   disk_used, disk_available, network_rx, network_tx, load_avg, process_count, timestamp
		FROM server_history
		WHERE hostname = ?
		  AND (project_key = ? OR ? = '')
		  AND timestamp >= ?
		  AND timestamp <= ?
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, hostname, projectKey, projectKey, start, end)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"hostname":    hostname,
			"project_key": projectKey,
			"start_time":  start,
			"end_time":    end,
		}).Error("Failed to query history by time range")
		return nil, fmt.Errorf("failed to query history by time range: %w", err)
	}
	defer rows.Close()

	var historyData []*models.HistoryData
	for rows.Next() {
		data := &models.HistoryData{}
		err := rows.Scan(
			&data.SessionID,
			&data.Hostname,
			&data.ProjectKey,
			&data.CPUUsage,
			&data.MemoryUsed,
			&data.MemoryAvailable,
			&data.DiskUsed,
			&data.DiskAvailable,
			&data.NetworkRx,
			&data.NetworkTx,
			&data.LoadAvg,
			&data.ProcessCount,
			&data.Timestamp,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan history row")
			continue
		}

		historyData = append(historyData, data)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating history rows")
		return nil, fmt.Errorf("error iterating history rows: %w", err)
	}

	return historyData, nil
}

// CleanupOldData 清理旧数据
func (r *SQLiteHistoryRepository) CleanupOldData(ctx context.Context, before time.Time) error {
	query := `DELETE FROM server_history WHERE timestamp < ?`

	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		r.logger.WithError(err).WithField("before", before).Error("Failed to cleanup old data")
		return fmt.Errorf("failed to cleanup old data: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.logger.WithFields(map[string]interface{}{
			"rows_affected": rowsAffected,
			"before":        before,
		}).Info("Old history data cleaned up")
	}

	return nil
}

// GetHistoryCount 获取历史数据数量
func (r *SQLiteHistoryRepository) GetHistoryCount(ctx context.Context, hostname, projectKey string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM server_history
		WHERE hostname = ? AND (project_key = ? OR ? = '')
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, hostname, projectKey, projectKey).Scan(&count)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"hostname":    hostname,
			"project_key": projectKey,
		}).Error("Failed to count history records")
		return 0, fmt.Errorf("failed to count history records: %w", err)
	}

	return count, nil
}

// GetAggregatedData 获取聚合数据
func (r *SQLiteHistoryRepository) GetAggregatedData(ctx context.Context, hostname, projectKey string, interval time.Duration, limit int) ([]*models.HistoryData, error) {
	// 根据时间间隔确定聚合函数
	var intervalStr string
	switch {
	case interval < time.Hour:
		intervalStr = "strftime('%Y-%m-%d %H:%M:00', timestamp)"
	case interval < 24*time.Hour:
		intervalStr = "strftime('%Y-%m-%d %H:00:00', timestamp)"
	default:
		intervalStr = "strftime('%Y-%m-%d', timestamp)"
	}

	query := fmt.Sprintf(`
		SELECT
			%s as time_bucket,
			AVG(cpu_usage) as cpu_usage,
			AVG(memory_used) as memory_used,
			AVG(memory_available) as memory_available,
			AVG(disk_used) as disk_used,
			AVG(disk_available) as disk_available,
			SUM(network_rx) as network_rx,
			SUM(network_tx) as network_tx,
			AVG(load_avg) as load_avg,
			AVG(process_count) as process_count,
			MIN(timestamp) as timestamp
		FROM server_history
		WHERE hostname = ?
		  AND (project_key = ? OR ? = '')
		  AND timestamp >= datetime('now', '-%d days')
		GROUP BY time_bucket
		ORDER BY time_bucket DESC
		LIMIT ?
	`, intervalStr, int(interval.Hours()*7)/24) // 获取大约7倍间隔时间的数据

	rows, err := r.db.QueryContext(ctx, query, hostname, projectKey, projectKey, limit)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"hostname":    hostname,
			"project_key": projectKey,
			"interval":    interval,
		}).Error("Failed to query aggregated data")
		return nil, fmt.Errorf("failed to query aggregated data: %w", err)
	}
	defer rows.Close()

	var aggregatedData []*models.HistoryData
	for rows.Next() {
		data := &models.HistoryData{}
		err := rows.Scan(
			&data.TimeBucket,
			&data.CPUUsage,
			&data.MemoryUsed,
			&data.MemoryAvailable,
			&data.DiskUsed,
			&data.DiskAvailable,
			&data.NetworkRx,
			&data.NetworkTx,
			&data.LoadAvg,
			&data.ProcessCount,
			&data.Timestamp,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan aggregated row")
			continue
		}

		aggregatedData = append(aggregatedData, data)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating aggregated rows")
		return nil, fmt.Errorf("error iterating aggregated rows: %w", err)
	}

	return aggregatedData, nil
}