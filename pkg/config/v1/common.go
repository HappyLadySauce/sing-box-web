package v1

import "time"

// DatabaseConfig defines database configuration
type DatabaseConfig struct {
	Driver   string `yaml:"driver" json:"driver"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	MaxIdleConns int           `yaml:"maxIdleConns" json:"maxIdleConns"`
	MaxOpenConns int           `yaml:"maxOpenConns" json:"maxOpenConns"`
	MaxLifetime  time.Duration `yaml:"maxLifetime" json:"maxLifetime"`
}

// LogConfig defines logging configuration
type LogConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	MaxSize    int    `yaml:"maxSize" json:"maxSize"`
	MaxAge     int    `yaml:"maxAge" json:"maxAge"`
	MaxBackups int    `yaml:"maxBackups" json:"maxBackups"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// APIServerConnection defines API server connection config
type APIServerConnection struct {
	Address  string        `yaml:"address" json:"address"`
	Port     int           `yaml:"port" json:"port"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Insecure bool          `yaml:"insecure" json:"insecure"`
	CertFile string        `yaml:"certFile" json:"certFile"`
	KeyFile  string        `yaml:"keyFile" json:"keyFile"`
	CAFile   string        `yaml:"caFile" json:"caFile"`
}

// MetricsConfig defines metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Address string `yaml:"address" json:"address"`
	Port    int    `yaml:"port" json:"port"`
	Path    string `yaml:"path" json:"path"`
}

// SkyWalkingConfig defines SkyWalking agent configuration
type SkyWalkingConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Collector   string `yaml:"collector" json:"collector"`
	ServiceName string `yaml:"serviceName" json:"serviceName"`
	SampleRate  int    `yaml:"sampleRate" json:"sampleRate"`
}