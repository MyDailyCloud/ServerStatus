package service

import (
	"context"
	"time"

	"github.com/kanshan/ServerStatus/data-server/internal/models"
)

// Service 业务服务层接口
type Service interface {
	ServerService
	ExportService
	AuthService
	HealthService
	ConfigService
}

// ServerService 服务器管理服务接口
type ServerService interface {
	// 服务器管理
	RegisterServer(ctx context.Context, info *models.SystemInfo) error
	UpdateServerStatus(ctx context.Context, info *models.SystemInfo) error
	GetServer(ctx context.Context, sessionID string) (*models.ServerInfo, error)
	GetServers(ctx context.Context, filter *ServerFilter) ([]*models.ServerInfo, error)
	RemoveServer(ctx context.Context, sessionID string) error

	// 查询操作
	GetServersByProject(ctx context.Context, projectKey string, pagination *Pagination) ([]*models.ServerInfo, error)
	GetServersByHostname(ctx context.Context, hostname string) ([]*models.ServerInfo, error)
	GetOnlineServers(ctx context.Context, projectKey string) ([]*models.ServerInfo, error)

	// 统计操作
	GetServerCount(ctx context.Context, projectKey string) (int, error)
	GetProjectStats(ctx context.Context, projectKey string) (*ProjectStats, error)

	// 状态管理
	UpdateServerOnlineStatus(ctx context.Context, sessionID string, online bool) error
	MarkServerAsOffline(ctx context.Context, timeout time.Duration) error
}

// ExportService 数据导出服务接口
type ExportService interface {
	// 导出操作
	ExportServers(ctx context.Context, req *ExportRequest) (*ExportResult, error)
	ExportHistory(ctx context.Context, req *ExportRequest) (*ExportResult, error)
	ExportUserResources(ctx context.Context, req *ExportRequest) (*ExportResult, error)

	// 格式支持
	SupportsFormat(format string) bool
	GetSupportedFormats() []string

	// 导出配置
	ValidateExportRequest(req *ExportRequest) error
	GetExportPreview(ctx context.Context, req *ExportRequest) (*ExportPreview, error)
}

// AuthService 认证服务接口
type AuthService interface {
	// 认证操作
	ValidateServerKey(ctx context.Context, key string) error
	GenerateAccessKey(ctx context.Context, serverKey, projectKey string) (string, error)
	ValidateAccessKey(ctx context.Context, accessKey string) (string, error)

	// 权限检查
	CheckProjectAccess(ctx context.Context, projectKey, accessKey string) error
	ValidateRequest(ctx context.Context, req *AuthRequest) (*AuthResult, error)

	// 密钥管理
	InvalidateAccessKey(ctx context.Context, accessKey string) error
	RefreshAccessKey(ctx context.Context, accessKey string) (string, error)
}

// HealthService 健康检查服务接口
type HealthService interface {
	// 健康检查
	CheckHealth(ctx context.Context) (*HealthStatus, error)
	CheckComponentHealth(ctx context.Context, component string) (*ComponentHealth, error)

	// 系统状态
	GetSystemStats(ctx context.Context) (*SystemStats, error)
	GetUptime() time.Duration
	GetVersion() string

	// 监控指标
	GetMetrics(ctx context.Context) (map[string]interface{}, error)
	IsHealthy() bool
}

// ConfigService 配置管理服务接口
type ConfigService interface {
	// 配置操作
	GetConfig() interface{}
	ReloadConfig() error
	ValidateConfig(config interface{}) error

	// 动态配置
	UpdateConfig(ctx context.Context, config interface{}) error
	WatchConfigChanges(ctx context.Context) (<-chan interface{}, error)

	// 配置历史
	GetConfigHistory() []ConfigVersion
	RestoreConfig(version string) error
}

// WebSocketService WebSocket服务接口
type WebSocketService interface {
	// 连接管理
	AcceptConnection(ctx context.Context, conn *WebSocketConnection, projectKey string) error
	CloseConnection(ctx context.Context, connID string) error
	GetConnections() []*WebSocketConnection

	// 消息广播
	BroadcastToProject(ctx context.Context, projectKey string, message interface{}) error
	BroadcastToAll(ctx context.Context, message interface{}) error
	SendToConnection(ctx context.Context, connID string, message interface{}) error

	// 统计信息
	GetConnectionStats() *WebSocketStats
	IsConnected(connID string) bool
}

// 数据结构定义

// ServerFilter 服务器查询过滤器
type ServerFilter struct {
	ProjectKey     string     `json:"project_key"`
	Hostname       string     `json:"hostname"`
	Status         string     `json:"status"` // online, offline, all
	OS             string     `json:"os"`
	Tags           []string   `json:"tags"`
	LastSeenAfter  *time.Time `json:"last_seen_after"`
	LastSeenBefore *time.Time `json:"last_seen_before"`
}

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
	Limit    int `json:"limit"`
}

// ProjectStats 项目统计信息
type ProjectStats struct {
	TotalServers   int       `json:"total_servers"`
	OnlineServers  int       `json:"online_servers"`
	OfflineServers int       `json:"offline_servers"`
	AvgCPUUsage    float64   `json:"avg_cpu_usage"`
	AvgMemoryUsage float64   `json:"avg_memory_usage"`
	AvgDiskUsage   float64   `json:"avg_disk_usage"`
	LastUpdateTime time.Time `json:"last_update_time"`
}

// ExportRequest 导出请求
type ExportRequest struct {
	Format     string            `json:"format"` // csv, json, xml
	Type       string            `json:"type"`   // servers, history, user_resources
	ProjectKey string            `json:"project_key"`
	Hostname   string            `json:"hostname"`
	Filters    map[string]string `json:"filters"`
	Limit      int               `json:"limit"`
	Fields     []string          `json:"fields"`
}

// ExportResult 导出结果
type ExportResult struct {
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	Format      string    `json:"format"`
	RecordCount int       `json:"record_count"`
	CreatedAt   time.Time `json:"created_at"`
	DownloadURL string    `json:"download_url"`
}

// ExportPreview 导出预览
type ExportPreview struct {
	Headers       []string   `json:"headers"`
	SampleRows    [][]string `json:"sample_rows"`
	TotalRows     int        `json:"total_rows"`
	EstimatedSize int64      `json:"estimated_size"`
}

// AuthRequest 认证请求
type AuthRequest struct {
	Type        string `json:"type"` // server_key, access_key
	Credentials string `json:"credentials"`
	ProjectKey  string `json:"project_key"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
}

// AuthResult 认证结果
type AuthResult struct {
	Valid      bool       `json:"valid"`
	ProjectKey string     `json:"project_key"`
	Role       string     `json:"role"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Message    string     `json:"message"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status     string                     `json:"status"` // healthy, degraded, unhealthy
	Uptime     time.Duration              `json:"uptime"`
	Version    string                     `json:"version"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status  string                 `json:"status"` // healthy, degraded, unhealthy
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

// SystemStats 系统统计
type SystemStats struct {
	CPUUsage    float64          `json:"cpu_usage"`
	MemoryUsage float64          `json:"memory_usage"`
	DiskUsage   float64          `json:"disk_usage"`
	Goroutines  int              `json:"goroutines"`
	HeapAlloc   uint64           `json:"heap_alloc"`
	Connections int              `json:"connections"`
	Requests    map[string]int64 `json:"requests"`
	Timestamp   time.Time        `json:"timestamp"`
}

// ConfigVersion 配置版本
type ConfigVersion struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	Changes   []string  `json:"changes"`
}

// WebSocketConnection WebSocket连接
type WebSocketConnection struct {
	ID          string    `json:"id"`
	ProjectKey  string    `json:"project_key"`
	RemoteAddr  string    `json:"remote_addr"`
	UserAgent   string    `json:"user_agent"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPing    time.Time `json:"last_ping"`
}

// WebSocketStats WebSocket统计信息
type WebSocketStats struct {
	TotalConnections  int     `json:"total_connections"`
	ActiveConnections int     `json:"active_connections"`
	TotalMessages     int64   `json:"total_messages"`
	MessagesPerSecond float64 `json:"messages_per_second"`
}
