// Package config provides configuration management for sing-box-web.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Manager 配置管理器
type Manager struct {
	v    *viper.Viper
	path string
}

// NewManager 创建新的配置管理器
func NewManager() *Manager {
	v := viper.New()
	
	// 设置配置文件类型
	v.SetConfigType("yaml")
	
	// 设置环境变量前缀
	v.SetEnvPrefix("SINGBOX")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	return &Manager{
		v: v,
	}
}

// LoadFromFile 从文件加载配置
func (m *Manager) LoadFromFile(configPath string) error {
	// 设置配置文件路径
	m.path = configPath
	
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}
	
	// 设置配置文件路径和名称
	dir := filepath.Dir(configPath)
	filename := filepath.Base(configPath)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	m.v.AddConfigPath(dir)
	m.v.SetConfigName(name)
	
	// 读取配置文件
	if err := m.v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	return nil
}

// LoadFromPaths 从多个路径尝试加载配置
func (m *Manager) LoadFromPaths(paths []string) error {
	var lastErr error
	
	for _, path := range paths {
		if err := m.LoadFromFile(path); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	
	return fmt.Errorf("failed to load config from any path: %w", lastErr)
}

// SetDefaults 设置默认值
func (m *Manager) SetDefaults(defaults map[string]interface{}) {
	for key, value := range defaults {
		m.v.SetDefault(key, value)
	}
}

// Get 获取配置值
func (m *Manager) Get(key string) interface{} {
	return m.v.Get(key)
}

// GetString 获取字符串配置值
func (m *Manager) GetString(key string) string {
	return m.v.GetString(key)
}

// GetInt 获取整数配置值
func (m *Manager) GetInt(key string) int {
	return m.v.GetInt(key)
}

// GetBool 获取布尔配置值
func (m *Manager) GetBool(key string) bool {
	return m.v.GetBool(key)
}

// GetDuration 获取时间间隔配置值
func (m *Manager) GetDuration(key string) time.Duration {
	return m.v.GetDuration(key)
}

// GetStringSlice 获取字符串数组配置值
func (m *Manager) GetStringSlice(key string) []string {
	return m.v.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置值
func (m *Manager) GetStringMap(key string) map[string]interface{} {
	return m.v.GetStringMap(key)
}

// GetStringMapString 获取字符串到字符串的映射配置值
func (m *Manager) GetStringMapString(key string) map[string]string {
	return m.v.GetStringMapString(key)
}

// Set 设置配置值
func (m *Manager) Set(key string, value interface{}) {
	m.v.Set(key, value)
}

// IsSet 检查配置项是否已设置
func (m *Manager) IsSet(key string) bool {
	return m.v.IsSet(key)
}

// AllKeys 获取所有配置键
func (m *Manager) AllKeys() []string {
	return m.v.AllKeys()
}

// AllSettings 获取所有配置
func (m *Manager) AllSettings() map[string]interface{} {
	return m.v.AllSettings()
}

// Unmarshal 将配置解析到结构体
func (m *Manager) Unmarshal(rawVal interface{}) error {
	return m.v.Unmarshal(rawVal)
}

// UnmarshalKey 将指定键的配置解析到结构体
func (m *Manager) UnmarshalKey(key string, rawVal interface{}) error {
	return m.v.UnmarshalKey(key, rawVal)
}

// WatchConfig 监听配置文件变化
func (m *Manager) WatchConfig() {
	m.v.WatchConfig()
}

// OnConfigChange 设置配置变化回调
func (m *Manager) OnConfigChange(run func(in fsnotify.Event)) {
	m.v.OnConfigChange(run)
}

// GetConfigFile 获取当前使用的配置文件路径
func (m *Manager) GetConfigFile() string {
	return m.v.ConfigFileUsed()
}

// WriteConfig 写入配置到文件
func (m *Manager) WriteConfig() error {
	return m.v.WriteConfig()
}

// WriteConfigAs 写入配置到指定文件
func (m *Manager) WriteConfigAs(filename string) error {
	return m.v.WriteConfigAs(filename)
}

// Sub 获取子配置
func (m *Manager) Sub(key string) *Manager {
	sub := m.v.Sub(key)
	if sub == nil {
		return nil
	}
	
	return &Manager{
		v:    sub,
		path: m.path,
	}
}

// DefaultConfigPaths 默认配置文件搜索路径
func DefaultConfigPaths(appName string) []string {
	return []string{
		fmt.Sprintf("./configs/%s/config.yaml", appName),
		fmt.Sprintf("/etc/%s/config.yaml", appName),
		fmt.Sprintf("$HOME/.%s/config.yaml", appName),
		"./config.yaml",
	}
}

// LoadConfig 便捷函数：加载指定应用的配置
func LoadConfig(appName string, configPath string) (*Manager, error) {
	manager := NewManager()
	
	// 设置默认值
	setDefaultValues(manager, appName)
	
	// 如果指定了配置文件路径，则使用指定路径
	if configPath != "" {
		return manager, manager.LoadFromFile(configPath)
	}
	
	// 否则从默认路径搜索
	paths := DefaultConfigPaths(appName)
	return manager, manager.LoadFromPaths(paths)
}

// setDefaultValues 设置默认配置值
func setDefaultValues(manager *Manager, appName string) {
	defaults := map[string]interface{}{
		"logging.level":  "info",
		"logging.format": "json",
		"logging.output": "stdout",
	}
	
	// 根据应用类型设置不同的默认值
	switch appName {
	case "api":
		defaults["server.host"] = "0.0.0.0"
		defaults["server.port"] = 8080
		defaults["grpc.host"] = "0.0.0.0"
		defaults["grpc.port"] = 9090
		defaults["metrics.enabled"] = true
		defaults["metrics.port"] = 9091
	case "web":
		defaults["server.host"] = "0.0.0.0"
		defaults["server.port"] = 8080
		defaults["static.embed_enabled"] = true
		defaults["static.spa_enabled"] = true
	case "agent":
		defaults["agent.manager.endpoint"] = "localhost:9090"
		defaults["agent.heartbeat.interval"] = "30s"
		defaults["monitoring.enabled"] = true
		defaults["monitoring.interval"] = "60s"
	}
	
	manager.SetDefaults(defaults)
} 