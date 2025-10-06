package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
	"github.com/kanshan/ServerStatus/data-server/pkg/utils"
	"gopkg.in/yaml.v3"
)

// Loader 配置加载器接口
type Loader interface {
	LoadFromFile(path string) (*Config, error)
	LoadFromBytes(data []byte) (*Config, error)
	SaveToFile(config *Config, path string) error
	GetConfigPaths() []string
}

// YAMLLoader YAML配置加载器
type YAMLLoader struct {
	configPaths []string
}

// NewYAMLLoader 创建YAML配置加载器
func NewYAMLLoader(configPaths []string) *YAMLLoader {
	return &YAMLLoader{
		configPaths: configPaths,
	}
}

// LoadFromFile 从文件加载配置
func (l *YAMLLoader) LoadFromFile(path string) (*Config, error) {
	// 检查文件是否存在
	if !utils.FileExists(path) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// 读取文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 加载配置
	config, err := l.LoadFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置配置文件路径
	config.ConfigPath = path

	return config, nil
}

// LoadFromBytes 从字节数据加载配置
func (l *YAMLLoader) LoadFromBytes(data []byte) (*Config, error) {
	var config Config

	// 解析YAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// 应用默认值
	config.ApplyDefaults()

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 处理环境变量替换
	if err := l.processEnvironmentVariables(&config); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	return &config, nil
}

// SaveToFile 保存配置到文件
func (l *YAMLLoader) SaveToFile(config *Config, path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := utils.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化为YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPaths 获取配置文件搜索路径
func (l *YAMLLoader) GetConfigPaths() []string {
	return l.configPaths
}

// LoadWithFallback 加载配置，支持多路径回退
func (l *YAMLLoader) LoadWithFallback(filenames []string) (*Config, error) {
	var lastErr error

	// 遍历所有可能的配置文件路径
	for _, filename := range filenames {
		for _, configDir := range l.configPaths {
			configPath := filepath.Join(configDir, filename)

			config, err := l.LoadFromFile(configPath)
			if err == nil {
				return config, nil
			}

			lastErr = err
		}
	}

	return nil, fmt.Errorf("no valid config file found, last error: %w", lastErr)
}

// processEnvironmentVariables 处理环境变量替换
func (l *YAMLLoader) processEnvironmentVariables(config *Config) error {
	// 数据库配置
	if strings.HasPrefix(config.Database.Path, "${") && strings.HasSuffix(config.Database.Path, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Database.Path, "}"), "${")
		config.Database.Path = utils.GetEnv(envKey, config.Database.Path)
	}

	// 缓存配置
	if strings.HasPrefix(config.Cache.Address, "${") && strings.HasSuffix(config.Cache.Address, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Cache.Address, "}"), "${")
		config.Cache.Address = utils.GetEnv(envKey, config.Cache.Address)
	}

	if strings.HasPrefix(config.Cache.Password, "${") && strings.HasSuffix(config.Cache.Password, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Cache.Password, "}"), "${")
		config.Cache.Password = utils.GetEnv(envKey, config.Cache.Password)
	}

	// 服务器配置
	if strings.HasPrefix(config.Server.Host, "${") && strings.HasSuffix(config.Server.Host, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Server.Host, "}"), "${}")
		config.Server.Host = utils.GetEnv(envKey, config.Server.Host)
	}

	// 日志配置
	if strings.HasPrefix(config.Logging.Filename, "${") && strings.HasSuffix(config.Logging.Filename, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Logging.Filename, "}"), "${}")
		config.Logging.Filename = utils.GetEnv(envKey, config.Logging.Filename)
	}

	// WebSocket配置
	if strings.HasPrefix(config.WebSocket.Path, "${") && strings.HasSuffix(config.WebSocket.Path, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.WebSocket.Path, "}"), "${}")
		config.WebSocket.Path = utils.GetEnv(envKey, config.WebSocket.Path)
	}

	return nil
}

// LoadDefaultConfig 加载默认配置
func LoadDefaultConfig() (*Config, error) {
	config := DefaultConfig()

	// 验证默认配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("default config validation failed: %w", err)
	}

	return config, nil
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() (*Config, error) {
	config := DefaultConfig()

	// 从环境变量覆盖配置
	if serverHost := utils.GetEnv("SERVER_HOST", ""); serverHost != "" {
		config.Server.Host = serverHost
	}

	if serverPort := utils.GetEnv("SERVER_PORT", ""); serverPort != "" {
		config.Server.Port = serverPort
	}

	if dbPath := utils.GetEnv("DB_PATH", ""); dbPath != "" {
		config.Database.Path = dbPath
	}

	if redisAddr := utils.GetEnv("REDIS_ADDR", ""); redisAddr != "" {
		config.Cache.Address = redisAddr
	}

	if redisPassword := utils.GetEnv("REDIS_PASSWORD", ""); redisPassword != "" {
		config.Cache.Password = redisPassword
	}

	if logLevel := utils.GetEnv("LOG_LEVEL", ""); logLevel != "" {
		config.Logging.Level = logger.Level(logLevel)
	}

	if logOutput := utils.GetEnv("LOG_OUTPUT", ""); logOutput != "" {
		config.Logging.Output = logOutput
	}

	if logFile := utils.GetEnv("LOG_FILE", ""); logFile != "" {
		config.Logging.Filename = logFile
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("env config validation failed: %w", err)
	}

	return config, nil
}
