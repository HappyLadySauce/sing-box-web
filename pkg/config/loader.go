package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	configv1 "sing-box-web/pkg/config/v1"
	"sing-box-web/pkg/config/validation"
)

// LoaderOptions defines options for configuration loading
type LoaderOptions struct {
	// ConfigPath is the path to the configuration file
	ConfigPath string

	// UseDefaults indicates whether to use default values for missing configurations
	UseDefaults bool

	// ValidateConfig indicates whether to validate the loaded configuration
	ValidateConfig bool

	// RequireFile indicates whether the configuration file must exist
	RequireFile bool
}

// Loader provides configuration loading functionality
type Loader struct {
	options LoaderOptions
}

// NewLoader creates a new configuration loader
func NewLoader(options LoaderOptions) *Loader {
	return &Loader{
		options: options,
	}
}

// LoadWebConfig loads web service configuration
func (l *Loader) LoadWebConfig() (*configv1.WebConfig, error) {
	var config *configv1.WebConfig

	// Start with defaults if enabled
	if l.options.UseDefaults {
		config = configv1.DefaultWebConfig()
	} else {
		config = &configv1.WebConfig{}
	}

	// Load from file if specified
	if l.options.ConfigPath != "" {
		if err := l.loadFromFile(l.options.ConfigPath, config); err != nil {
			if l.options.RequireFile || !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load web config from file: %w", err)
			}
		}
	}

	// Validate configuration if enabled
	if l.options.ValidateConfig {
		if err := validation.ValidateWebConfig(config); err != nil {
			return nil, fmt.Errorf("web config validation failed: %w", err)
		}
	}

	return config, nil
}

// LoadAPIConfig loads API service configuration
func (l *Loader) LoadAPIConfig() (*configv1.APIConfig, error) {
	var config *configv1.APIConfig

	// Start with defaults if enabled
	if l.options.UseDefaults {
		config = configv1.DefaultAPIConfig()
	} else {
		config = &configv1.APIConfig{}
	}

	// Load from file if specified
	if l.options.ConfigPath != "" {
		if err := l.loadFromFile(l.options.ConfigPath, config); err != nil {
			if l.options.RequireFile || !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load API config from file: %w", err)
			}
		}
	}

	// Validate configuration if enabled
	if l.options.ValidateConfig {
		if err := validation.ValidateAPIConfig(config); err != nil {
			return nil, fmt.Errorf("API config validation failed: %w", err)
		}
	}

	return config, nil
}

// LoadAgentConfig loads agent service configuration
func (l *Loader) LoadAgentConfig() (*configv1.AgentConfig, error) {
	var config *configv1.AgentConfig

	// Start with defaults if enabled
	if l.options.UseDefaults {
		config = configv1.DefaultAgentConfig()
	} else {
		config = &configv1.AgentConfig{}
	}

	// Load from file if specified
	if l.options.ConfigPath != "" {
		if err := l.loadFromFile(l.options.ConfigPath, config); err != nil {
			if l.options.RequireFile || !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load agent config from file: %w", err)
			}
		}
	}

	// Validate configuration if enabled
	if l.options.ValidateConfig {
		if err := validation.ValidateAgentConfig(config); err != nil {
			return nil, fmt.Errorf("agent config validation failed: %w", err)
		}
	}

	return config, nil
}

// loadFromFile loads configuration from a file
func (l *Loader) loadFromFile(path string, config interface{}) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	// Read file content
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Determine file format based on extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return l.loadFromYAML(data, config)
	case ".json":
		return l.loadFromJSON(data, config)
	default:
		// Default to YAML if extension is unknown
		return l.loadFromYAML(data, config)
	}
}

// loadFromYAML loads configuration from YAML data
func (l *Loader) loadFromYAML(data []byte, config interface{}) error {
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}
	return nil
}

// loadFromJSON loads configuration from JSON data
func (l *Loader) loadFromJSON(data []byte, config interface{}) error {
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse JSON config: %w", err)
	}
	return nil
}

// SaveWebConfig saves web configuration to file
func SaveWebConfig(config *configv1.WebConfig, path string) error {
	return saveConfig(config, path)
}

// SaveAPIConfig saves API configuration to file
func SaveAPIConfig(config *configv1.APIConfig, path string) error {
	return saveConfig(config, path)
}

// SaveAgentConfig saves agent configuration to file
func SaveAgentConfig(config *configv1.AgentConfig, path string) error {
	return saveConfig(config, path)
}

// saveConfig saves configuration to file
func saveConfig(config interface{}, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ViperLoader provides Viper-based configuration loading
type ViperLoader struct {
	v *viper.Viper
}

// NewViperLoader creates a new Viper-based configuration loader
func NewViperLoader(configName, configPath string, envPrefix string) *ViperLoader {
	v := viper.New()

	// Set configuration file name and path
	if configName != "" {
		v.SetConfigName(configName)
	}
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/sing-box/")

	// Set environment variable prefix
	if envPrefix != "" {
		v.SetEnvPrefix(envPrefix)
	}
	v.AutomaticEnv()

	// Replace dots with underscores in environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return &ViperLoader{v: v}
}

// LoadConfig loads configuration using Viper
func (vl *ViperLoader) LoadConfig(config interface{}) error {
	// Try to read configuration file
	if err := vl.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration
	if err := vl.v.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// SetDefault sets a default value for a configuration key
func (vl *ViperLoader) SetDefault(key string, value interface{}) {
	vl.v.SetDefault(key, value)
}

// SetOverride sets an override value for a configuration key
func (vl *ViperLoader) SetOverride(key string, value interface{}) {
	vl.v.Set(key, value)
}

// GetConfigFilePath returns the path of the configuration file used
func (vl *ViperLoader) GetConfigFilePath() string {
	return vl.v.ConfigFileUsed()
}

// MergeConfigs merges environment variables and command line flags into configuration
func MergeConfigs(config interface{}, envPrefix string, overrides map[string]interface{}) error {
	// Create a Viper loader for merging
	vl := NewViperLoader("", "", envPrefix)

	// Apply overrides
	for key, value := range overrides {
		vl.SetOverride(key, value)
	}

	// Load configuration with environment variables and overrides
	return vl.LoadConfig(config)
}

// GetConfigTemplate returns a template configuration for the specified service
func GetConfigTemplate(service string) (interface{}, error) {
	switch service {
	case "web":
		return configv1.DefaultWebConfig(), nil
	case "api":
		return configv1.DefaultAPIConfig(), nil
	case "agent":
		return configv1.DefaultAgentConfig(), nil
	default:
		return nil, fmt.Errorf("unknown service: %s", service)
	}
}

// ValidateConfigFile validates a configuration file without loading it into a struct
func ValidateConfigFile(path string, service string) error {
	loader := NewLoader(LoaderOptions{
		ConfigPath:     path,
		UseDefaults:    false,
		ValidateConfig: true,
		RequireFile:    true,
	})

	switch service {
	case "web":
		_, err := loader.LoadWebConfig()
		return err
	case "api":
		_, err := loader.LoadAPIConfig()
		return err
	case "agent":
		_, err := loader.LoadAgentConfig()
		return err
	default:
		return fmt.Errorf("unknown service: %s", service)
	}
}
