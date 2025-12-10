package main

import "time"

// DBStore 抽象数据库接口，便于未来替换存储实现（SQLite / PostgreSQL / MySQL 等）
// 当前由 database.go 的 Database 实现。
type DBStore interface {
	SaveServerInfo(sessionID, hostname, projectKey string, latestData *SystemInfo) error
	SaveHistoryData(sessionID, hostname, projectKey string, data *SystemInfo) error

	GetServerInfo(sessionID string) (*ServerInfo, error)
	GetHistoryData(sessionID string, limit int) ([]*SystemInfo, error)
	GetHistoryByTimeRange(sessionID, projectKey string, from, to time.Time, limit int) ([]*SystemInfo, error)
	GetAllServers(projectKey string, offset, limit int) ([]*ServerInfo, error)
	GetServerCount(projectKey string) (int, error)

	CleanupOldData(olderThan time.Duration, maxHistoryPerServer int) error

	SaveAccessKeyCache(cacheKey, accessKey string) error
	GetAccessKeyCache(cacheKey string) (string, error)

	GetUUIDStats() (map[string]interface{}, error)

	SaveVisitorEvent(event *VisitorEvent) error
	GetVisitorStats(projectKey string, since time.Time) (*VisitorStats, error)
	GetVisitorAggregation(projectKey, groupBy string, since time.Time, limit int) ([]AggregatedVisitorItem, error)

	UpsertDomainBinding(binding DomainBinding) error
	GetProjectKeyByDomain(domain string) (string, error)
	ListDomainBindings() ([]DomainBinding, error)

	SaveOrUpdateUser(user *User) (int64, error)
	CreateSession(userID int64, token string, expiresAt time.Time) error
	GetSession(token string) (*Session, error)
	DeleteSession(token string) error
	GetUserByID(id int64) (*User, error)
	UpsertUserConfig(userID int64, config string) error
	GetUserConfig(userID int64) (string, error)

	Ping() error
	Close() error
}
