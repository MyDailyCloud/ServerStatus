package config

import (
	"fmt"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
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
	// TODO: 实现JSON配置加载
	return nil, fmt.Errorf("JSON loader not implemented yet")
}

// LoadFromBytes 从字节数据加载JSON配置
func (l *JSONLoader) LoadFromBytes(data []byte) (*Config, error) {
	// TODO: 实现JSON配置加载
	return nil, fmt.Errorf("JSON loader not implemented yet")
}

// SaveToFile 保存JSON配置到文件
func (l *JSONLoader) SaveToFile(config *Config, path string) error {
	// TODO: 实现JSON配置保存
	return fmt.Errorf("JSON loader not implemented yet")
}

// GetConfigPaths 获取配置文件搜索路径
func (l *JSONLoader) GetConfigPaths() []string {
	return l.configPaths
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
