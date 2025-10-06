package config

import (
	"time"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

// Config 应用配置
type Config struct {
	ConfigPath string          `yaml:"-" json:"config_path"`
	Server     ServerConfig    `yaml:"server" json:"server"`
	Database   DatabaseConfig  `yaml:"database" json:"database"`
	Cache      CacheConfig     `yaml:"cache" json:"cache"`
	Logging    LoggingConfig   `yaml:"logging" json:"logging"`
	WebSocket  WebSocketConfig `yaml:"websocket" json:"websocket"`
	Export     ExportConfig    `yaml:"export" json:"export"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string   `yaml:"host" json:"host"`
	Port           string   `yaml:"port" json:"port"`
	ReadTimeout    string   `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout   string   `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout    string   `yaml:"idle_timeout" json:"idle_timeout"`
	MaxMemoryMB    int      `yaml:"max_memory_mb" json:"max_memory_mb"`
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
	RequireAuth    bool     `yaml:"require_auth" json:"require_auth"`
	ProjectKey     string   `yaml:"project_key" json:"project_key"`
	ServerKey      string   `yaml:"server_key" json:"server_key"`
	DataLimit      int      `yaml:"data_limit" json:"data_limit"`
	DataInterval   int      `yaml:"data_interval" json:"data_interval"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type             string `yaml:"type" json:"type"`
	Path             string `yaml:"path" json:"path"`
	ConnectionString string `yaml:"connection_string" json:"connection_string"`
	DSN              string `yaml:"dsn" json:"dsn"`
	MaxOpenConns     int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns     int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime  string `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Type          string `yaml:"type" json:"type"` // redis, memory, none
	Address       string `yaml:"address" json:"address"`
	Password      string `yaml:"password" json:"password"`
	DB            int    `yaml:"db" json:"db"`
	PoolSize      int    `yaml:"pool_size" json:"pool_size"`
	MinIdleConns  int    `yaml:"min_idle_conns" json:"min_idle_conns"`
	DialTimeout   string `yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout   string `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout  string `yaml:"write_timeout" json:"write_timeout"`
	DefaultTTL    string `yaml:"default_ttl" json:"default_ttl"`
	MaxMemoryMB   int    `yaml:"max_memory_mb" json:"max_memory_mb"`
	Enable        bool   `yaml:"enable" json:"enable"`
	RedisAddr     string `yaml:"redis_addr" json:"redis_addr"`
	RedisPassword string `yaml:"redis_password" json:"redis_password"`
	RedisDB       int    `yaml:"redis_db" json:"redis_db"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level        logger.Level  `yaml:"level" json:"level"`       // debug, info, warn, error
	Format       logger.Format `yaml:"format" json:"format"`     // json, text
	Output       string        `yaml:"output" json:"output"`     // stdout, file, both
	Filename     string        `yaml:"filename" json:"filename"` // 日志文件名
	MaxSize      int           `yaml:"max_size" json:"max_size"` // MB
	MaxBackups   int           `yaml:"max_backups" json:"max_backups"`
	MaxAge       int           `yaml:"max_age" json:"max_age"` // days
	Compress     bool          `yaml:"compress" json:"compress"`
	ReportCaller bool          `yaml:"report_caller" json:"report_caller"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	Enabled           bool   `yaml:"enabled" json:"enabled"`
	Path              string `yaml:"path" json:"path"`
	Port              string `yaml:"port" json:"port"`
	ReadBufferSize    int    `yaml:"read_buffer_size" json:"read_buffer_size"`
	WriteBufferSize   int    `yaml:"write_buffer_size" json:"write_buffer_size"`
	PingInterval      string `yaml:"ping_interval" json:"ping_interval"`
	PongWait          string `yaml:"pong_wait" json:"pong_wait"`
	WriteWait         string `yaml:"write_wait" json:"write_wait"`
	MaxMessageSize    int64  `yaml:"max_message_size" json:"max_message_size"`
	EnableCompression bool   `yaml:"enable_compression" json:"enable_compression"`
}

// ExportConfig 导出配置
type ExportConfig struct {
	Directory        string   `yaml:"directory" json:"directory"`
	RetentionDays    int      `yaml:"retention_days" json:"retention_days"`
	MaxFileSize      int64    `yaml:"max_file_size" json:"max_file_size"`
	SupportedFormats []string `yaml:"supported_formats" json:"supported_formats"`
	Enable           bool     `yaml:"enable" json:"enable"`
	Formats          []string `yaml:"formats" json:"formats"`       // csv, json, xml
	MaxRows          int      `yaml:"max_rows" json:"max_rows"`     // 单次导出最大行数
	ChunkSize        int      `yaml:"chunk_size" json:"chunk_size"` // 分片大小
	TempDir          string   `yaml:"temp_dir" json:"temp_dir"`     // 临时文件目录
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:           "0.0.0.0",
			Port:           "8080",
			AllowedOrigins: []string{"*"},
			RequireAuth:    true,
			ProjectKey:     "public",
			ServerKey:      "serverstatus.ltd",
			DataLimit:      1000,
			DataInterval:   5,
		},
		Database: DatabaseConfig{
			Type:            "sqlite",
			Path:            "/tmp/serverstatus.db",
			DSN:             "./data/serverstatus.db",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: "5m",
		},
		Cache: CacheConfig{
			Enable:      false,
			Type:        "memory",
			RedisAddr:   "localhost:6379",
			RedisDB:     0,
			DefaultTTL:  "30m",
			MaxMemoryMB: 100,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
		WebSocket: WebSocketConfig{
			Enabled:           true,
			Path:              "/ws",
			Port:              "8080",
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			PingInterval:      "54s",
			PongWait:          "60s",
			WriteWait:         "10s",
			MaxMessageSize:    512,
			EnableCompression: false,
		},
		Export: ExportConfig{
			Directory:        "/tmp/exports",
			RetentionDays:    7,
			MaxFileSize:      104857600,
			SupportedFormats: []string{"csv", "json", "xml"},
			Enable:           true,
			Formats:          []string{"csv", "json"},
			MaxRows:          10000,
			ChunkSize:        1000,
			TempDir:          "/tmp",
		},
	}
}

// GetAddr 获取服务器地址
func (c *ServerConfig) GetAddr() string {
	return c.Host + ":" + c.Port
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Logging.Level == "info" || c.Logging.Level == "warn" || c.Logging.Level == "error"
}

// GetDefaultTTL 获取默认TTL
func (c *CacheConfig) GetDefaultTTL() time.Duration {
	if c.DefaultTTL == "" {
		return 30 * time.Minute
	}
	dur, err := time.ParseDuration(c.DefaultTTL)
	if err != nil {
		return 30 * time.Minute
	}
	return dur
}

// ApplyDefaults 应用默认配置值
func (c *Config) ApplyDefaults() {
	// 服务器默认配置
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if len(c.Server.AllowedOrigins) == 0 {
		c.Server.AllowedOrigins = []string{"*"}
	}
	if c.Server.ReadTimeout == "" {
		c.Server.ReadTimeout = "30s"
	}
	if c.Server.WriteTimeout == "" {
		c.Server.WriteTimeout = "30s"
	}
	if c.Server.IdleTimeout == "" {
		c.Server.IdleTimeout = "60s"
	}

	// 数据库默认配置
	if c.Database.Type == "" {
		c.Database.Type = "sqlite"
	}
	if c.Database.Path == "" && c.Database.Type == "sqlite" {
		c.Database.Path = "./data/serverstatus.db"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.ConnMaxLifetime == "" {
		c.Database.ConnMaxLifetime = "5m"
	}

	// 缓存默认配置
	if c.Cache.Type == "" {
		c.Cache.Type = "memory"
	}
	if c.Cache.Address == "" && c.Cache.Type == "redis" {
		c.Cache.Address = "localhost:6379"
	}
	if c.Cache.PoolSize == 0 {
		c.Cache.PoolSize = 10
	}
	if c.Cache.MinIdleConns == 0 {
		c.Cache.MinIdleConns = 3
	}
	if c.Cache.DialTimeout == "" {
		c.Cache.DialTimeout = "5s"
	}
	if c.Cache.ReadTimeout == "" {
		c.Cache.ReadTimeout = "3s"
	}
	if c.Cache.WriteTimeout == "" {
		c.Cache.WriteTimeout = "3s"
	}
	if c.Cache.DefaultTTL == "" {
		c.Cache.DefaultTTL = "30m"
	}

	// 日志默认配置
	if c.Logging.Level == "" {
		c.Logging.Level = logger.InfoLevel
	}
	if c.Logging.Format == "" {
		c.Logging.Format = logger.TextFormat
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = 100
	}
	if c.Logging.MaxAge == 0 {
		c.Logging.MaxAge = 30
	}

	// WebSocket默认配置
	if c.WebSocket.Path == "" {
		c.WebSocket.Path = "/ws"
	}
	if c.WebSocket.Port == "" {
		c.WebSocket.Port = "8080"
	}
	if c.WebSocket.ReadBufferSize == 0 {
		c.WebSocket.ReadBufferSize = 1024
	}
	if c.WebSocket.WriteBufferSize == 0 {
		c.WebSocket.WriteBufferSize = 1024
	}
	if c.WebSocket.PingInterval == "" {
		c.WebSocket.PingInterval = "30s"
	}
	if c.WebSocket.PongWait == "" {
		c.WebSocket.PongWait = "60s"
	}
	if c.WebSocket.WriteWait == "" {
		c.WebSocket.WriteWait = "10s"
	}
	if c.WebSocket.MaxMessageSize == 0 {
		c.WebSocket.MaxMessageSize = 1048576
	}

	// 导出默认配置
	if c.Export.Directory == "" {
		c.Export.Directory = "./exports"
	}
	if c.Export.MaxFileSize == 0 {
		c.Export.MaxFileSize = 104857600 // 100MB
	}
	if len(c.Export.SupportedFormats) == 0 {
		c.Export.SupportedFormats = []string{"csv", "json", "xml"}
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	validator := NewConfigValidator()
	return validator.Validate(c)
}
