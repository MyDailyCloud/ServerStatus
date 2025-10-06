package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kanshan/ServerStatus/data-server/pkg/utils"
)

// Validator 配置验证器接口
type Validator interface {
	Validate(config *Config) error
	ValidateServerConfig(config ServerConfig) error
	ValidateDatabaseConfig(config DatabaseConfig) error
	ValidateCacheConfig(config CacheConfig) error
	ValidateLoggingConfig(config LoggingConfig) error
	ValidateWebSocketConfig(config WebSocketConfig) error
	ValidateExportConfig(config ExportConfig) error
}

// ConfigValidator 配置验证器实现
type ConfigValidator struct{}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// Validate 验证完整配置
func (v *ConfigValidator) Validate(config *Config) error {
	if err := v.ValidateServerConfig(config.Server); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}

	if err := v.ValidateDatabaseConfig(config.Database); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}

	if err := v.ValidateCacheConfig(config.Cache); err != nil {
		return fmt.Errorf("cache config validation failed: %w", err)
	}

	if err := v.ValidateLoggingConfig(config.Logging); err != nil {
		return fmt.Errorf("logging config validation failed: %w", err)
	}

	if err := v.ValidateWebSocketConfig(config.WebSocket); err != nil {
		return fmt.Errorf("websocket config validation failed: %w", err)
	}

	if err := v.ValidateExportConfig(config.Export); err != nil {
		return fmt.Errorf("export config validation failed: %w", err)
	}

	return nil
}

// ValidateServerConfig 验证服务器配置
func (v *ConfigValidator) ValidateServerConfig(config ServerConfig) error {
	// 验证主机地址
	if config.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}

	if !utils.IsValidHost(config.Host) {
		return fmt.Errorf("invalid server host: %s", config.Host)
	}

	// 验证端口号
	if !utils.IsValidPort(config.Port) {
		return fmt.Errorf("invalid server port: %s", config.Port)
	}

	// 验证读取超时
	if config.ReadTimeout != "" {
		if _, err := utils.ParseDuration(config.ReadTimeout); err != nil {
			return fmt.Errorf("invalid read timeout: %s", config.ReadTimeout)
		}
	}

	// 验证写入超时
	if config.WriteTimeout != "" {
		if _, err := utils.ParseDuration(config.WriteTimeout); err != nil {
			return fmt.Errorf("invalid write timeout: %s", config.WriteTimeout)
		}
	}

	// 验证空闲超时
	if config.IdleTimeout != "" {
		if _, err := utils.ParseDuration(config.IdleTimeout); err != nil {
			return fmt.Errorf("invalid idle timeout: %s", config.IdleTimeout)
		}
	}

	// 验证CORS配置
	if len(config.AllowedOrigins) == 0 {
		return fmt.Errorf("allowed origins cannot be empty")
	}

	for _, origin := range config.AllowedOrigins {
		if origin != "*" {
			if !v.isValidOrigin(origin) {
				return fmt.Errorf("invalid allowed origin: %s", origin)
			}
		}
	}

	return nil
}

// ValidateDatabaseConfig 验证数据库配置
func (v *ConfigValidator) ValidateDatabaseConfig(config DatabaseConfig) error {
	// 验证数据库类型
	supportedTypes := []string{"sqlite", "mysql", "postgresql"}
	if !utils.Contains(supportedTypes, config.Type) {
		return fmt.Errorf("unsupported database type: %s, supported types: %v", config.Type, supportedTypes)
	}

	// 验证连接字符串或文件路径
	switch config.Type {
	case "sqlite":
		if config.Path == "" {
			return fmt.Errorf("sqlite database path cannot be empty")
		}

		if !filepath.IsAbs(config.Path) {
			return fmt.Errorf("sqlite database path must be absolute: %s", config.Path)
		}

		// 确保目录存在
		dir := filepath.Dir(config.Path)
		if !utils.DirExists(dir) {
			if err := utils.EnsureDir(dir); err != nil {
				return fmt.Errorf("failed to create database directory: %w", err)
			}
		}

	case "mysql", "postgresql":
		if config.ConnectionString == "" {
			return fmt.Errorf("connection string cannot be empty for %s", config.Type)
		}

		// 验证连接字符串格式
		if !v.isValidConnectionString(config.ConnectionString, config.Type) {
			return fmt.Errorf("invalid connection string for %s", config.Type)
		}
	}

	// 验证最大连接数
	if config.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be greater than 0")
	}

	if config.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}

	if config.MaxIdleConns > config.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot be greater than max open connections")
	}

	// 验证连接超时
	if config.ConnMaxLifetime != "" {
		if _, err := utils.ParseDuration(config.ConnMaxLifetime); err != nil {
			return fmt.Errorf("invalid connection max lifetime: %s", config.ConnMaxLifetime)
		}
	}

	return nil
}

// ValidateCacheConfig 验证缓存配置
func (v *ConfigValidator) ValidateCacheConfig(config CacheConfig) error {
	// 验证缓存类型
	supportedTypes := []string{"redis", "memory", "none"}
	if !utils.Contains(supportedTypes, config.Type) {
		return fmt.Errorf("unsupported cache type: %s, supported types: %v", config.Type, supportedTypes)
	}

	if config.Type == "none" {
		return nil
	}

	// 验证Redis配置
	if config.Type == "redis" {
		if config.Address == "" {
			return fmt.Errorf("redis address cannot be empty")
		}

		if !utils.IsValidRedisAddr(config.Address) {
			return fmt.Errorf("invalid redis address: %s", config.Address)
		}

		// 验证数据库索引
		if config.DB < 0 || config.DB > 15 {
			return fmt.Errorf("redis database index must be between 0 and 15")
		}

		// 验证连接池大小
		if config.PoolSize <= 0 {
			return fmt.Errorf("redis pool size must be greater than 0")
		}

		// 验证超时设置
		if config.DialTimeout != "" {
			if _, err := utils.ParseDuration(config.DialTimeout); err != nil {
				return fmt.Errorf("invalid redis dial timeout: %s", config.DialTimeout)
			}
		}

		if config.ReadTimeout != "" {
			if _, err := utils.ParseDuration(config.ReadTimeout); err != nil {
				return fmt.Errorf("invalid redis read timeout: %s", config.ReadTimeout)
			}
		}

		if config.WriteTimeout != "" {
			if _, err := utils.ParseDuration(config.WriteTimeout); err != nil {
				return fmt.Errorf("invalid redis write timeout: %s", config.WriteTimeout)
			}
		}
	}

	return nil
}

// ValidateLoggingConfig 验证日志配置
func (v *ConfigValidator) ValidateLoggingConfig(config LoggingConfig) error {
	// 验证日志级别
	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !utils.Contains(validLevels, string(config.Level)) {
		return fmt.Errorf("invalid log level: %s, valid levels: %v", config.Level, validLevels)
	}

	// 验证日志格式
	validFormats := []string{"json", "text"}
	if !utils.Contains(validFormats, string(config.Format)) {
		return fmt.Errorf("invalid log format: %s, valid formats: %v", config.Format, validFormats)
	}

	// 验证输出方式
	validOutputs := []string{"stdout", "file", "both"}
	if !utils.Contains(validOutputs, config.Output) {
		return fmt.Errorf("invalid log output: %s, valid outputs: %v", config.Output, validOutputs)
	}

	// 验证文件路径
	if config.Output == "file" || config.Output == "both" {
		if config.Filename == "" {
			return fmt.Errorf("log filename cannot be empty when output is file or both")
		}

		if !filepath.IsAbs(config.Filename) {
			return fmt.Errorf("log filename must be absolute path: %s", config.Filename)
		}

		// 确保目录存在
		dir := filepath.Dir(config.Filename)
		if !utils.DirExists(dir) {
			if err := utils.EnsureDir(dir); err != nil {
				return fmt.Errorf("failed to create log directory: %w", err)
			}
		}

		// 验证文件大小限制
		if config.MaxSize <= 0 {
			return fmt.Errorf("log max size must be greater than 0")
		}

		// 验证备份文件数量
		if config.MaxBackups < 0 {
			return fmt.Errorf("log max backups cannot be negative")
		}

		// 验证保留时间
		if config.MaxAge < 0 {
			return fmt.Errorf("log max age cannot be negative")
		}
	}

	return nil
}

// ValidateWebSocketConfig 验证WebSocket配置
func (v *ConfigValidator) ValidateWebSocketConfig(config WebSocketConfig) error {
	// 验证路径
	if config.Path == "" {
		return fmt.Errorf("websocket path cannot be empty")
	}

	if !strings.HasPrefix(config.Path, "/") {
		return fmt.Errorf("websocket path must start with /")
	}

	// 验证缓冲区大小
	if config.ReadBufferSize <= 0 {
		return fmt.Errorf("websocket read buffer size must be greater than 0")
	}

	if config.WriteBufferSize <= 0 {
		return fmt.Errorf("websocket write buffer size must be greater than 0")
	}

	// 验证Ping间隔
	if config.PingInterval != "" {
		if _, err := utils.ParseDuration(config.PingInterval); err != nil {
			return fmt.Errorf("invalid websocket ping interval: %s", config.PingInterval)
		}
	}

	// 验证Pong等待时间
	if config.PongWait != "" {
		if _, err := utils.ParseDuration(config.PongWait); err != nil {
			return fmt.Errorf("invalid websocket pong wait: %s", config.PongWait)
		}
	}

	// 验证写入等待时间
	if config.WriteWait != "" {
		if _, err := utils.ParseDuration(config.WriteWait); err != nil {
			return fmt.Errorf("invalid websocket write wait: %s", config.WriteWait)
		}
	}

	// 验证最大消息大小
	if config.MaxMessageSize <= 0 {
		return fmt.Errorf("websocket max message size must be greater than 0")
	}

	return nil
}

// ValidateExportConfig 验证导出配置
func (v *ConfigValidator) ValidateExportConfig(config ExportConfig) error {
	// 验证导出目录
	if config.Directory == "" {
		return fmt.Errorf("export directory cannot be empty")
	}

	if !filepath.IsAbs(config.Directory) {
		return fmt.Errorf("export directory must be absolute path: %s", config.Directory)
	}

	// 确保目录存在
	if !utils.DirExists(config.Directory) {
		if err := utils.EnsureDir(config.Directory); err != nil {
			return fmt.Errorf("failed to create export directory: %w", err)
		}
	}

	// 验证文件保留时间
	if config.RetentionDays < 0 {
		return fmt.Errorf("export retention days cannot be negative")
	}

	// 验证最大文件大小
	if config.MaxFileSize <= 0 {
		return fmt.Errorf("export max file size must be greater than 0")
	}

	// 验证支持格式
	validFormats := []string{"csv", "json", "xml", "excel"}
	for _, format := range config.SupportedFormats {
		if !utils.Contains(validFormats, format) {
			return fmt.Errorf("unsupported export format: %s, supported formats: %v", format, validFormats)
		}
	}

	if len(config.SupportedFormats) == 0 {
		return fmt.Errorf("at least one export format must be supported")
	}

	return nil
}

// isValidOrigin 验证CORS源地址
func (v *ConfigValidator) isValidOrigin(origin string) bool {
	// 支持http/https URL
	if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
		if _, err := url.Parse(origin); err != nil {
			return false
		}
		return true
	}

	// 支持通配符域名
	if strings.Contains(origin, "*") {
		pattern := strings.ReplaceAll(regexp.QuoteMeta(origin), "\\*", ".*")
		matched, _ := regexp.MatchString("^"+pattern+"$", "example.com")
		return matched
	}

	return false
}

// isValidConnectionString 验证连接字符串格式
func (v *ConfigValidator) isValidConnectionString(connStr, dbType string) bool {
	switch dbType {
	case "mysql":
		// MySQL连接字符串格式: user:password@tcp(host:port)/dbname
		mysqlPattern := `^[^:]+:[^@]+@tcp\([^:]+:\d+\)/[^/]+$`
		matched, _ := regexp.MatchString(mysqlPattern, connStr)
		return matched

	case "postgresql":
		// PostgreSQL连接字符串格式: postgres://user:password@host:port/dbname
		postgresPattern := `^postgres://[^:]+:[^@]+@[^:]+:\d+/[^/]+$`
		matched, _ := regexp.MatchString(postgresPattern, connStr)
		return matched
	}

	return false
}

// ValidateAdvanced 高级配置验证
func (v *ConfigValidator) ValidateAdvanced(config *Config) error {
	// 验证端口冲突
	if config.Server.Port == config.WebSocket.Port && config.WebSocket.Path != "" {
		return fmt.Errorf("server port and websocket port cannot be the same")
	}

	// 验证内存限制
	if config.Server.MaxMemoryMB > 0 && config.Server.MaxMemoryMB < 64 {
		return fmt.Errorf("max memory should be at least 64MB")
	}

	// 验证文件权限
	if config.Logging.Filename != "" {
		// 检查日志文件是否可写
		dir := filepath.Dir(config.Logging.Filename)
		if utils.DirExists(dir) {
			file := filepath.Join(dir, ".test_write")
			if f, err := os.Create(file); err != nil {
				return fmt.Errorf("log directory is not writable: %w", err)
			} else {
				f.Close()
				os.Remove(file)
			}
		}
	}

	return nil
}
