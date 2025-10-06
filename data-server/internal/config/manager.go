package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
	"github.com/kanshan/ServerStatus/data-server/pkg/utils"
	"gopkg.in/yaml.v3"
)

// Manager 配置管理器接口
type Manager interface {
	Load() (*Config, error)
	Reload() error
	GetConfig() *Config
	Watch(ctx context.Context) (<-chan *Config, error)
	Stop()
	UpdateConfig(updates func(*Config) error) error
	ValidateConfig(config *Config) error
	GetVersion() string
}

// ConfigManager 配置管理器实现
type ConfigManager struct {
	config      *Config
	configPath  string
	loader      Loader
	validator   Validator
	logger      logger.Logger
	mu          sync.RWMutex
	version     string
	lastModTime time.Time
	stopCh      chan struct{}
	watchers    []chan *Config
	watchersMu  sync.Mutex
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string, logger logger.Logger) Manager {
	return &ConfigManager{
		configPath: configPath,
		loader:     NewYAMLLoader([]string{"/etc/serverstatus", "/usr/local/etc/serverstatus", "."}),
		validator:  NewConfigValidator(),
		logger:     logger,
		version:    utils.GenerateUUID(),
		stopCh:     make(chan struct{}),
		watchers:   make([]chan *Config, 0),
	}
}

// Load 加载配置
func (m *ConfigManager) Load() (*Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var config *Config
	var err error

	// 尝试从指定路径加载
	if m.configPath != "" && utils.FileExists(m.configPath) {
		config, err = m.loader.LoadFromFile(m.configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", m.configPath, err)
		}
	} else {
		// 尝试从默认路径加载
		config, err = m.loader.(*YAMLLoader).LoadWithFallback([]string{"config.yaml", "config.yml", "serverstatus.yaml"})
		if err != nil {
			m.logger.Warnf("Failed to load config from file, using default config: %v", err)
			config, err = LoadDefaultConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to load default config: %w", err)
			}
		}
	}

	// 从环境变量覆盖配置
	envConfig, err := LoadConfigFromEnv()
	if err != nil {
		m.logger.Warnf("Failed to load env config: %v", err)
	} else {
		// 合并环境变量配置
		m.mergeConfig(config, envConfig)
	}

	// 最终验证
	if err := m.validator.Validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	m.config = config
	m.lastModTime = time.Now()
	m.version = utils.GenerateUUID()

	m.logger.Infof("Config loaded successfully from: %s", config.ConfigPath)
	return config, nil
}

// Reload 重新加载配置
func (m *ConfigManager) Reload() error {
	m.logger.Info("Reloading configuration...")

	newConfig, err := m.Load()
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// 通知所有观察者
	m.notifyWatchers(newConfig)

	m.logger.Info("Configuration reloaded successfully")
	return nil
}

// GetConfig 获取当前配置
func (m *ConfigManager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return nil
	}

	// 返回配置的副本
	configCopy := *m.config
	return &configCopy
}

// Watch 监听配置变化
func (m *ConfigManager) Watch(ctx context.Context) (<-chan *Config, error) {
	m.watchersMu.Lock()
	defer m.watchersMu.Unlock()

	ch := make(chan *Config, 1)
	m.watchers = append(m.watchers, ch)

	// 启动文件监听
	go m.watchConfigFile(ctx)

	return ch, nil
}

// Stop 停止配置管理器
func (m *ConfigManager) Stop() {
	close(m.stopCh)

	m.watchersMu.Lock()
	defer m.watchersMu.Unlock()

	// 关闭所有观察者通道
	for _, watcher := range m.watchers {
		close(watcher)
	}
	m.watchers = make([]chan *Config, 0)
}

// UpdateConfig 更新配置
func (m *ConfigManager) UpdateConfig(updates func(*Config) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return fmt.Errorf("config not loaded")
	}

	// 创建配置副本
	configCopy := *m.config

	// 应用更新
	if err := updates(&configCopy); err != nil {
		return fmt.Errorf("failed to apply config updates: %w", err)
	}

	// 验证新配置
	if err := m.validator.Validate(&configCopy); err != nil {
		return fmt.Errorf("updated config validation failed: %w", err)
	}

	// 保存到文件
	if m.configPath != "" {
		if err := m.loader.SaveToFile(&configCopy, m.configPath); err != nil {
			return fmt.Errorf("failed to save updated config: %w", err)
		}
	}

	// 更新内存配置
	m.config = &configCopy
	m.version = utils.GenerateUUID()

	// 通知观察者
	m.notifyWatchers(&configCopy)

	m.logger.Info("Configuration updated successfully")
	return nil
}

// ValidateConfig 验证配置
func (m *ConfigManager) ValidateConfig(config *Config) error {
	return m.validator.Validate(config)
}

// GetVersion 获取配置版本
func (m *ConfigManager) GetVersion() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.version
}

// mergeConfig 合并配置
func (m *ConfigManager) mergeConfig(target, source *Config) {
	if source.Server.Host != "" {
		target.Server.Host = source.Server.Host
	}
	if source.Server.Port != "" {
		target.Server.Port = source.Server.Port
	}

	if source.Database.Type != "" {
		target.Database.Type = source.Database.Type
	}
	if source.Database.Path != "" {
		target.Database.Path = source.Database.Path
	}
	if source.Database.ConnectionString != "" {
		target.Database.ConnectionString = source.Database.ConnectionString
	}

	if source.Cache.Type != "" {
		target.Cache.Type = source.Cache.Type
	}
	if source.Cache.Address != "" {
		target.Cache.Address = source.Cache.Address
	}

	if source.Logging.Level != "" {
		target.Logging.Level = source.Logging.Level
	}
	if source.Logging.Output != "" {
		target.Logging.Output = source.Logging.Output
	}
	if source.Logging.Filename != "" {
		target.Logging.Filename = source.Logging.Filename
	}
}

// watchConfigFile 监听配置文件变化
func (m *ConfigManager) watchConfigFile(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			if m.configPath != "" && utils.FileExists(m.configPath) {
				if info, err := utils.GetFileSize(m.configPath); err == nil {
					modTime := time.Unix(info, 0)
					if modTime.After(m.lastModTime) {
						m.logger.Info("Config file changed, reloading...")
						if err := m.Reload(); err != nil {
							m.logger.Errorf("Failed to reload config: %v", err)
						}
					}
				}
			}
		}
	}
}

// notifyWatchers 通知所有观察者
func (m *ConfigManager) notifyWatchers(config *Config) {
	m.watchersMu.Lock()
	defer m.watchersMu.Unlock()

	for _, watcher := range m.watchers {
		select {
		case watcher <- config:
		default:
			m.logger.Warn("Config watcher channel is full, skipping notification")
		}
	}
}

// GetConfigSummary 获取配置摘要信息
func (m *ConfigManager) GetConfigSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return map[string]interface{}{
			"loaded": false,
		}
	}

	return map[string]interface{}{
		"loaded":            true,
		"version":           m.version,
		"config_path":       m.config.ConfigPath,
		"server_host":       m.config.Server.Host,
		"server_port":       m.config.Server.Port,
		"database_type":     m.config.Database.Type,
		"cache_type":        m.config.Cache.Type,
		"log_level":         m.config.Logging.Level,
		"websocket_enabled": m.config.WebSocket.Enabled,
		"last_updated":      m.lastModTime.Format(time.RFC3339),
	}
}

// ValidateAdvanced 高级配置验证
func (m *ConfigManager) ValidateAdvanced() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return fmt.Errorf("config not loaded")
	}

	validator := m.validator.(*ConfigValidator)
	return validator.ValidateAdvanced(m.config)
}

// ExportConfig 导出当前配置
func (m *ConfigManager) ExportConfig(format string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	switch format {
	case "yaml", "yml":
		// 使用已有的loader来序列化
		if buf, err := yaml.Marshal(m.config); err != nil {
			return nil, fmt.Errorf("failed to marshal config to YAML: %w", err)
		} else {
			return buf, nil
		}

	case "json":
		if buf, err := json.MarshalIndent(m.config, "", "  "); err != nil {
			return nil, fmt.Errorf("failed to marshal config to JSON: %w", err)
		} else {
			return buf, nil
		}

	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}
