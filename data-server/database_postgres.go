package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var pgMigrations embed.FS

// PostgresDatabase 实现 DBStore，支持 PostgreSQL
type PostgresDatabase struct {
	db *sql.DB
}

// NewPostgresDatabase 创建并迁移 Postgres 数据库
func NewPostgresDatabase(dsn string) (*PostgresDatabase, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// 连接验证
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	// 执行迁移
	if err := runPostgresMigrations(db); err != nil {
		return nil, fmt.Errorf("postgres migrations failed: %w", err)
	}

	return &PostgresDatabase{db: db}, nil
}

// runPostgresMigrations 运行嵌入的迁移
func runPostgresMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	d, err := iofs.New(pgMigrations, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return err
	}

	// 忽略已完成的 Up
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// SaveServerInfo UPSERT server_info
func (d *PostgresDatabase) SaveServerInfo(sessionID, hostname, projectKey string, latestData *SystemInfo) error {
	dataJSON, err := json.Marshal(latestData)
	if err != nil {
		return fmt.Errorf("序列化服务器信息失败: %w", err)
	}

	query := `
	INSERT INTO server_info (session_id, hostname, project_key, latest_data, last_seen)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (session_id)
	DO UPDATE SET hostname = EXCLUDED.hostname,
	              project_key = EXCLUDED.project_key,
	              latest_data = EXCLUDED.latest_data,
	              last_seen = EXCLUDED.last_seen,
	              updated_at = CURRENT_TIMESTAMP;
	`
	_, err = d.db.Exec(query, sessionID, hostname, projectKey, string(dataJSON), time.Now())
	if err != nil {
		return fmt.Errorf("保存服务器信息失败: %w", err)
	}
	return nil
}

// SaveHistoryData 写入历史
func (d *PostgresDatabase) SaveHistoryData(sessionID, hostname, projectKey string, data *SystemInfo) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化历史数据失败: %w", err)
	}
	query := `
	INSERT INTO history_data (session_id, hostname, project_key, timestamp, data)
	VALUES ($1, $2, $3, $4, $5);
	`
	_, err = d.db.Exec(query, sessionID, hostname, projectKey, data.Timestamp, string(dataJSON))
	if err != nil {
		return fmt.Errorf("保存历史数据失败: %w", err)
	}
	return nil
}

// GetServerInfo 读取最新
func (d *PostgresDatabase) GetServerInfo(sessionID string) (*ServerInfo, error) {
	var latestData string
	var lastSeen time.Time
	query := `SELECT latest_data, last_seen FROM server_info WHERE session_id = $1;`
	err := d.db.QueryRow(query, sessionID).Scan(&latestData, &lastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var systemInfo SystemInfo
	if err := json.Unmarshal([]byte(latestData), &systemInfo); err != nil {
		return nil, err
	}
	return &ServerInfo{
		Latest:   &systemInfo,
		LastSeen: lastSeen,
		History:  make([]*SystemInfo, 0),
	}, nil
}

// GetHistoryData 最近 N 条
func (d *PostgresDatabase) GetHistoryData(sessionID string, limit int) ([]*SystemInfo, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
	SELECT data, timestamp
	FROM history_data
	WHERE session_id = $1
	ORDER BY timestamp DESC
	LIMIT $2;
	`
	rows, err := d.db.Query(query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*SystemInfo
	for rows.Next() {
		var dataStr string
		var ts time.Time
		if err := rows.Scan(&dataStr, &ts); err != nil {
			continue
		}
		var info SystemInfo
		if err := json.Unmarshal([]byte(dataStr), &info); err != nil {
			continue
		}
		history = append(history, &info)
	}
	return history, nil
}

// GetHistoryByTimeRange 时间范围查询
func (d *PostgresDatabase) GetHistoryByTimeRange(sessionID, projectKey string, from, to time.Time, limit int) ([]*SystemInfo, error) {
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}
	where := "session_id = $1 AND timestamp BETWEEN $2 AND $3"
	args := []interface{}{sessionID, from, to, limit}
	argPos := 4
	if projectKey != "" {
		where = "session_id = $1 AND project_key = $2 AND timestamp BETWEEN $3 AND $4"
		args = []interface{}{sessionID, projectKey, from, to, limit}
		argPos = 5
	}

	query := fmt.Sprintf(`
	SELECT data, timestamp
	FROM history_data
	WHERE %s
	ORDER BY timestamp DESC
	LIMIT $%d;
	`, where, argPos)

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*SystemInfo
	for rows.Next() {
		var dataStr string
		var ts time.Time
		if err := rows.Scan(&dataStr, &ts); err != nil {
			continue
		}
		var info SystemInfo
		if err := json.Unmarshal([]byte(dataStr), &info); err != nil {
			continue
		}
		history = append(history, &info)
	}
	return history, nil
}

// GetAllServers 分页获取服务器
func (d *PostgresDatabase) GetAllServers(projectKey string, offset, limit int) ([]*ServerInfo, error) {
	if limit <= 0 {
		limit = 100
	}
	where := ""
	args := []interface{}{limit, offset}
	if projectKey != "" && projectKey != "public" {
		where = "WHERE project_key = $3"
		args = append(args, projectKey)
	}
	query := fmt.Sprintf(`
	SELECT session_id, hostname, latest_data, last_seen
	FROM server_info
	%s
	ORDER BY hostname
	LIMIT $1 OFFSET $2;
	`, where)

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*ServerInfo
	for rows.Next() {
		var sessionID, hostname, latestDataStr string
		var lastSeen time.Time
		if err := rows.Scan(&sessionID, &hostname, &latestDataStr, &lastSeen); err != nil {
			continue
		}
		var latest SystemInfo
		if err := json.Unmarshal([]byte(latestDataStr), &latest); err != nil {
			continue
		}
		servers = append(servers, &ServerInfo{
			Latest:   &latest,
			LastSeen: lastSeen,
			History:  make([]*SystemInfo, 0),
		})
	}
	return servers, nil
}

// GetServerCount 数量
func (d *PostgresDatabase) GetServerCount(projectKey string) (int, error) {
	var query string
	var count int
	if projectKey == "" || projectKey == "public" {
		query = "SELECT COUNT(*) FROM server_info;"
		err := d.db.QueryRow(query).Scan(&count)
		return count, err
	}
	query = "SELECT COUNT(*) FROM server_info WHERE project_key = $1;"
	err := d.db.QueryRow(query, projectKey).Scan(&count)
	return count, err
}

// CleanupOldData 清理历史
func (d *PostgresDatabase) CleanupOldData(olderThan time.Duration, maxHistoryPerServer int) error {
	cutoff := time.Now().Add(-olderThan)
	if _, err := d.db.Exec(`DELETE FROM history_data WHERE timestamp < $1`, cutoff); err != nil {
		return err
	}
	// 保留每个 session 最新 N 条
	_, err := d.db.Exec(`
	DELETE FROM history_data
	WHERE id IN (
		SELECT id FROM (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY timestamp DESC) AS rn
			FROM history_data
		) t WHERE t.rn > $1
	);`, maxHistoryPerServer)
	return err
}

// SaveAccessKeyCache
func (d *PostgresDatabase) SaveAccessKeyCache(cacheKey, accessKey string) error {
	_, err := d.db.Exec(`
	INSERT INTO access_key_cache (cache_key, access_key)
	VALUES ($1, $2)
	ON CONFLICT (cache_key) DO UPDATE SET access_key = EXCLUDED.access_key;
	`, cacheKey, accessKey)
	return err
}

func (d *PostgresDatabase) GetAccessKeyCache(cacheKey string) (string, error) {
	var accessKey string
	err := d.db.QueryRow(`SELECT access_key FROM access_key_cache WHERE cache_key = $1;`, cacheKey).Scan(&accessKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return accessKey, nil
}

// GetUUIDStats
func (d *PostgresDatabase) GetUUIDStats() (map[string]interface{}, error) {
	var totalServers, activeUUIDs, hostnameOnly int
	query := `
	SELECT
		COUNT(*) as total_servers,
		COUNT(CASE WHEN session_id != hostname THEN 1 END) as active_uuids,
		COUNT(CASE WHEN session_id = hostname THEN 1 END) as hostname_only
	FROM server_info;
	`
	if err := d.db.QueryRow(query).Scan(&totalServers, &activeUUIDs, &hostnameOnly); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_servers": totalServers,
		"active_uuids":  activeUUIDs,
		"hostname_only": hostnameOnly,
		"timestamp":     time.Now(),
		"description":   "使用我们服务的设备统计",
	}, nil
}

// Visitor events
func (d *PostgresDatabase) SaveVisitorEvent(event *VisitorEvent) error {
	if event == nil {
		return fmt.Errorf("访客事件为空")
	}
	_, err := d.db.Exec(`
	INSERT INTO visitor_events (project_key, domain, page_url, referrer, user_agent, ip, session_id, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8);
	`, event.ProjectKey, event.Domain, event.PageURL, event.Referrer, event.UserAgent, event.IP, event.SessionID, event.Timestamp)
	return err
}

func (d *PostgresDatabase) GetVisitorStats(projectKey string, since time.Time) (*VisitorStats, error) {
	if since.IsZero() {
		return nil, fmt.Errorf("统计时间范围无效")
	}
	where := "created_at >= $1"
	args := []interface{}{since}
	if projectKey != "" {
		where = "project_key = $1 AND created_at >= $2"
		args = []interface{}{projectKey, since}
	}

	// total
	var total int
	if err := d.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM visitor_events WHERE %s", where), args...).Scan(&total); err != nil {
		return nil, err
	}
	// unique IP
	var uniqueIPs int
	if err := d.db.QueryRow(fmt.Sprintf("SELECT COUNT(DISTINCT ip) FROM visitor_events WHERE %s", where), args...).Scan(&uniqueIPs); err != nil {
		return nil, err
	}
	// daily
	rows, err := d.db.Query(fmt.Sprintf(`
		SELECT TO_CHAR(created_at::date, 'YYYY-MM-DD') AS day, COUNT(*) AS cnt
		FROM visitor_events
		WHERE %s
		GROUP BY day
		ORDER BY day;
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var daily []DailyVisitorStat
	for rows.Next() {
		var stat DailyVisitorStat
		if err := rows.Scan(&stat.Date, &stat.Count); err != nil {
			continue
		}
		daily = append(daily, stat)
	}
	// top pages
	pageRows, err := d.db.Query(fmt.Sprintf(`
		SELECT page_url, COUNT(*) AS cnt
		FROM visitor_events
		WHERE %s AND page_url <> ''
		GROUP BY page_url
		ORDER BY cnt DESC
		LIMIT 10;
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer pageRows.Close()
	var topPages []PageVisitorStat
	for pageRows.Next() {
		var stat PageVisitorStat
		if err := pageRows.Scan(&stat.Page, &stat.Count); err != nil {
			continue
		}
		topPages = append(topPages, stat)
	}

	return &VisitorStats{
		ProjectKey:  projectKey,
		From:        since,
		To:          time.Now(),
		TotalVisits: total,
		UniqueIPs:   uniqueIPs,
		Daily:       daily,
		TopPages:    topPages,
	}, nil
}

func (d *PostgresDatabase) GetVisitorAggregation(projectKey, groupBy string, since time.Time, limit int) ([]AggregatedVisitorItem, error) {
	if since.IsZero() {
		return nil, fmt.Errorf("统计时间范围无效")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var column string
	switch groupBy {
	case "page":
		column = "page_url"
	case "referrer":
		column = "referrer"
	case "domain":
		column = "domain"
	case "ua", "user_agent":
		column = "user_agent"
	default:
		return nil, fmt.Errorf("不支持的分组字段: %s", groupBy)
	}

	where := "created_at >= $1"
	args := []interface{}{since}
	if projectKey != "" {
		where = "project_key = $1 AND created_at >= $2"
		args = []interface{}{projectKey, since}
	}
	query := fmt.Sprintf(`
	SELECT %s AS key, COUNT(*) AS cnt
	FROM visitor_events
	WHERE %s AND %s <> ''
	GROUP BY %s
	ORDER BY cnt DESC
	LIMIT $%d;
	`, column, where, column, column, len(args)+1)

	args = append(args, limit)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AggregatedVisitorItem
	for rows.Next() {
		var item AggregatedVisitorItem
		if err := rows.Scan(&item.Key, &item.Count); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// Domain binding
func (d *PostgresDatabase) UpsertDomainBinding(binding DomainBinding) error {
	if binding.Domain == "" || binding.ProjectKey == "" {
		return fmt.Errorf("域名或项目密钥不能为空")
	}
	_, err := d.db.Exec(`
	INSERT INTO domain_bindings (domain, project_key)
	VALUES ($1, $2)
	ON CONFLICT (domain) DO UPDATE SET project_key = EXCLUDED.project_key;
	`, binding.Domain, binding.ProjectKey)
	return err
}

func (d *PostgresDatabase) GetProjectKeyByDomain(domain string) (string, error) {
	if domain == "" {
		return "", nil
	}
	var projectKey string
	err := d.db.QueryRow(`SELECT project_key FROM domain_bindings WHERE domain = $1 LIMIT 1;`, domain).Scan(&projectKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return projectKey, nil
}

func (d *PostgresDatabase) ListDomainBindings() ([]DomainBinding, error) {
	rows, err := d.db.Query(`SELECT domain, project_key FROM domain_bindings ORDER BY domain;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []DomainBinding
	for rows.Next() {
		var b DomainBinding
		if err := rows.Scan(&b.Domain, &b.ProjectKey); err != nil {
			continue
		}
		list = append(list, b)
	}
	return list, nil
}

// User/session
func (d *PostgresDatabase) SaveOrUpdateUser(user *User) (int64, error) {
	if user == nil || user.GithubID == "" {
		return 0, fmt.Errorf("用户信息无效")
	}
	var id int64
	// upsert
	err := d.db.QueryRow(`
	INSERT INTO users (github_id, login, name, avatar_url, email)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (github_id) DO UPDATE SET
	  login = EXCLUDED.login,
	  name = EXCLUDED.name,
	  avatar_url = EXCLUDED.avatar_url,
	  email = EXCLUDED.email,
	  updated_at = CURRENT_TIMESTAMP
	RETURNING id;
	`, user.GithubID, user.Login, user.Name, user.AvatarURL, user.Email).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *PostgresDatabase) CreateSession(userID int64, token string, expiresAt time.Time) error {
	_, err := d.db.Exec(`
	INSERT INTO sessions (user_id, session_token, expires_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (session_token) DO NOTHING;
	`, userID, token, expiresAt)
	return err
}

func (d *PostgresDatabase) GetSession(token string) (*Session, error) {
	var s Session
	err := d.db.QueryRow(`
	SELECT session_token, user_id, expires_at, created_at
	FROM sessions
	WHERE session_token = $1;
	`, token).Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (d *PostgresDatabase) DeleteSession(token string) error {
	_, err := d.db.Exec(`DELETE FROM sessions WHERE session_token = $1;`, token)
	return err
}

func (d *PostgresDatabase) GetUserByID(id int64) (*User, error) {
	var u User
	err := d.db.QueryRow(`
	SELECT id, github_id, login, name, avatar_url, email, created_at, updated_at
	FROM users WHERE id = $1;
	`, id).Scan(&u.ID, &u.GithubID, &u.Login, &u.Name, &u.AvatarURL, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (d *PostgresDatabase) UpsertUserConfig(userID int64, config string) error {
	if config == "" {
		config = "{}"
	}
	_, err := d.db.Exec(`
	INSERT INTO user_configs (user_id, config)
	VALUES ($1, $2)
	ON CONFLICT (user_id) DO UPDATE SET config = EXCLUDED.config, updated_at = CURRENT_TIMESTAMP;
	`, userID, config)
	return err
}

func (d *PostgresDatabase) GetUserConfig(userID int64) (string, error) {
	var cfg string
	err := d.db.QueryRow(`SELECT config FROM user_configs WHERE user_id = $1;`, userID).Scan(&cfg)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return cfg, nil
}

func (d *PostgresDatabase) Close() error {
	return d.db.Close()
}

func (d *PostgresDatabase) Ping() error {
	return d.db.Ping()
}

