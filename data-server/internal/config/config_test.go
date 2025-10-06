package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kanshan/ServerStatus/data-server/pkg/logger"
)

func TestConfigValidation(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "Valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "Invalid server port",
			config: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: "invalid",
				},
				Database: DatabaseConfig{
					Type: "sqlite",
					Path: "/tmp/test.db",
				},
				Cache: CacheConfig{
					Type: "memory",
				},
				Logging: LoggingConfig{
					Level:  logger.InfoLevel,
					Format: logger.TextFormat,
					Output: "stdout",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid database type",
			config: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: "8080",
				},
				Database: DatabaseConfig{
					Type: "invalid",
					Path: "/tmp/test.db",
				},
				Cache: CacheConfig{
					Type: "memory",
				},
				Logging: LoggingConfig{
					Level:  logger.InfoLevel,
					Format: logger.TextFormat,
					Output: "stdout",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYAMLLoader(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	config := DefaultConfig()
	config.Server.Host = "test-host"
	config.Server.Port = "9090"

	loader := NewYAMLLoader([]string{tempDir})

	// 测试保存配置
	err := loader.SaveToFile(config, configPath)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// 测试加载配置
	loadedConfig, err := loader.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if loadedConfig.Server.Host != config.Server.Host {
		t.Errorf("Expected host %s, got %s", config.Server.Host, loadedConfig.Server.Host)
	}

	if loadedConfig.Server.Port != config.Server.Port {
		t.Errorf("Expected port %s, got %s", config.Server.Port, loadedConfig.Server.Port)
	}
}

func TestConfigManager(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "manager_test_config.yaml")

	// 创建测试日志器
	logConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}

	// 创建配置管理器
	manager := NewConfigManager(configPath, testLogger)

	// 测试加载配置
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	// 测试获取配置
	retrievedConfig := manager.GetConfig()
	if retrievedConfig == nil {
		t.Fatal("Expected non-nil retrieved config")
	}

	// 测试配置版本
	version := manager.GetVersion()
	if version == "" {
		t.Error("Expected non-empty version")
	}

	// 测试配置验证
	err = manager.ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v", err)
	}

	// 测试配置更新
	err = manager.UpdateConfig(func(c *Config) error {
		c.Server.Host = "updated-host"
		return nil
	})
	if err != nil {
		t.Errorf("UpdateConfig() error = %v", err)
	}

	// 验证配置已更新
	updatedConfig := manager.GetConfig()
	if updatedConfig.Server.Host != "updated-host" {
		t.Errorf("Expected updated host 'updated-host', got '%s'", updatedConfig.Server.Host)
	}
}

func TestConfigWatch(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "watch_test_config.yaml")

	// 创建测试日志器
	logConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}

	// 创建配置管理器
	manager := NewConfigManager(configPath, testLogger)

	// 加载初始配置
	_, err = manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 开始监听配置变化
	ch, err := manager.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch() error = %v", err)
	}

	// 修改配置文件
	go func() {
		time.Sleep(100 * time.Millisecond)
		manager.UpdateConfig(func(c *Config) error {
			c.Server.Host = "watched-host"
			return nil
		})
	}()

	// 等待配置变化通知
	select {
	case newConfig := <-ch:
		if newConfig.Server.Host != "watched-host" {
			t.Errorf("Expected host 'watched-host', got '%s'", newConfig.Server.Host)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for config change notification")
	}

	// 停止配置管理器
	manager.Stop()
}

func TestEnvironmentVariableOverrides(t *testing.T) {
	// 设置环境变量
	os.Setenv("SERVER_HOST", "env-host")
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("LOG_LEVEL")
	}()

	// 从环境变量加载配置
	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}

	if config.Server.Host != "env-host" {
		t.Errorf("Expected host 'env-host', got '%s'", config.Server.Host)
	}

	if config.Server.Port != "9999" {
		t.Errorf("Expected port '9999', got '%s'", config.Server.Port)
	}

	if config.Logging.Level != logger.DebugLevel {
		t.Errorf("Expected log level 'debug', got '%s'", config.Logging.Level)
	}
}

func TestFactory(t *testing.T) {
	factory := NewConfigFactory()

	// 测试创建验证器
	validator := factory.CreateValidator()
	if validator == nil {
		t.Error("Expected non-nil validator")
	}

	// 测试创建加载器
	yamlLoader := factory.CreateLoader("yaml")
	if yamlLoader == nil {
		t.Error("Expected non-nil YAML loader")
	}

	jsonLoader := factory.CreateLoader("json")
	if jsonLoader == nil {
		t.Error("Expected non-nil JSON loader")
	}

	// 测试创建管理器
	logConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.TextFormat,
		Output: "stdout",
	}
	testLogger, _ := logger.NewLogger(logConfig)

	manager := factory.CreateManager("", testLogger)
	if manager == nil {
		t.Error("Expected non-nil config manager")
	}
}

func TestConfigAdvancedValidation(t *testing.T) {
	validator := NewConfigValidator()

	// 测试高级验证
	config := DefaultConfig()

	// 设置有问题的配置
	config.Server.Port = "8080"
	config.WebSocket.Port = "8080" // 相同端口应该失败

	err := validator.ValidateAdvanced(config)
	if err == nil {
		t.Error("Expected validation error for port conflict")
	}

	// 修复端口冲突
	config.WebSocket.Port = "8081"

	err = validator.ValidateAdvanced(config)
	if err != nil {
		t.Errorf("Unexpected validation error after fix: %v", err)
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	validator := NewConfigValidator()
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(config)
	}
}

func BenchmarkConfigLoad(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "benchmark_config.yaml")

	config := DefaultConfig()
	loader := NewYAMLLoader([]string{tempDir})

	// 保存配置文件
	err := loader.SaveToFile(config, configPath)
	if err != nil {
		b.Fatalf("SaveToFile() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadFromFile(configPath)
		if err != nil {
			b.Fatalf("LoadFromFile() error = %v", err)
		}
	}
}
