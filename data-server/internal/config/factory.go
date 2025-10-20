package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
	"github.com/kanshan/ServerStatus/data-server/pkg/utils"
)

// Factory 配置工厂接口
type Factory interface {
	CreateManager(configPath string, logger logger.Logger) Manager
	CreateLoader(loaderType string) Loader
	CreateValidator() Validator
}

// ConfigFactory 配置工厂实现
type ConfigFactory struct{}

// NewConfigFactory 创建配置工厂
func NewConfigFactory() Factory {
	return &ConfigFactory{}
}

// CreateManager 创建配置管理器
func (f *ConfigFactory) CreateManager(configPath string, logger logger.Logger) Manager {
	return NewConfigManager(configPath, logger)
}

// CreateLoader 创建配置加载器
func (f *ConfigFactory) CreateLoader(loaderType string) Loader {
	switch loaderType {
	case "yaml", "yml":
		return NewYAMLLoader([]string{
			"/etc/serverstatus",
			"/usr/local/etc/serverstatus",
			".",
			"./config",
			"./configs",
		})
	case "json":
		return NewJSONLoader([]string{
			"/etc/serverstatus",
			"/usr/local/etc/serverstatus",
			".",
			"./config",
			"./configs",
		})
	default:
		return NewYAMLLoader([]string{
			"/etc/serverstatus",
			"/usr/local/etc/serverstatus",
			".",
		})
	}
}

// CreateValidator 创建配置验证器
func (f *ConfigFactory) CreateValidator() Validator {
	return NewConfigValidator()
}

// JSONLoader JSON配置加载器
type JSONLoader struct {
	configPaths []string
}

// NewJSONLoader 创建JSON配置加载器
func NewJSONLoader(configPaths []string) *JSONLoader {
	return &JSONLoader{
		configPaths: configPaths,
	}
}

// LoadFromFile 从文件加载JSON配置
func (l *JSONLoader) LoadFromFile(path string) (*Config, error) {
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

// LoadFromBytes 从字节数据加载JSON配置
func (l *JSONLoader) LoadFromBytes(data []byte) (*Config, error) {
	var config Config

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
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

// SaveToFile 保存JSON配置到文件
func (l *JSONLoader) SaveToFile(config *Config, path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := utils.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
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
func (l *JSONLoader) GetConfigPaths() []string {
	return l.configPaths
}

// processEnvironmentVariables 处理环境变量替换
func (l *JSONLoader) processEnvironmentVariables(config *Config) error {
	if strings.HasPrefix(config.Database.Path, "${") && strings.HasSuffix(config.Database.Path, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Database.Path, "}"), "${")
		config.Database.Path = utils.GetEnv(envKey, config.Database.Path)
	}

	if strings.HasPrefix(config.Cache.Address, "${") && strings.HasSuffix(config.Cache.Address, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Cache.Address, "}"), "${")
		config.Cache.Address = utils.GetEnv(envKey, config.Cache.Address)
	}

	if strings.HasPrefix(config.Cache.Password, "${") && strings.HasSuffix(config.Cache.Password, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Cache.Password, "}"), "${")
		config.Cache.Password = utils.GetEnv(envKey, config.Cache.Password)
	}

	if strings.HasPrefix(config.Server.Host, "${") && strings.HasSuffix(config.Server.Host, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Server.Host, "}"), "${}")
		config.Server.Host = utils.GetEnv(envKey, config.Server.Host)
	}

	if strings.HasPrefix(config.Logging.Filename, "${") && strings.HasSuffix(config.Logging.Filename, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.Logging.Filename, "}"), "${}")
		config.Logging.Filename = utils.GetEnv(envKey, config.Logging.Filename)
	}

	if strings.HasPrefix(config.WebSocket.Path, "${") && strings.HasSuffix(config.WebSocket.Path, "}") {
		envKey := strings.TrimPrefix(strings.TrimSuffix(config.WebSocket.Path, "}"), "${}")
		config.WebSocket.Path = utils.GetEnv(envKey, config.WebSocket.Path)
	}

	return nil
}

// 全局配置工厂实例
var DefaultFactory = NewConfigFactory()

// 便捷函数

// LoadConfig 加载配置的便捷函数
func LoadConfig(configPath string, logger logger.Logger) (*Config, error) {
	manager := DefaultFactory.CreateManager(configPath, logger)
	return manager.Load()
}

// ValidateConfig 验证配置的便捷函数
func ValidateConfig(config *Config) error {
	validator := DefaultFactory.CreateValidator()
	return validator.Validate(config)
}

// CreateConfigManager 创建配置管理器的便捷函数
func CreateConfigManager(configPath string, logger logger.Logger) Manager {
	return DefaultFactory.CreateManager(configPath, logger)
}
