package v1

import "time"

// WebConfig defines configuration for sing-box-web service
type WebConfig struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	
	// Server configuration
	Server ServerConfig `yaml:"server" json:"server"`
	
	// Database configuration
	Database DatabaseConfig `yaml:"database" json:"database"`
	
	// API server connection
	APIServer APIServerConnection `yaml:"apiServer" json:"apiServer"`
	
	// Authentication configuration
	Auth AuthConfig `yaml:"auth" json:"auth"`
	
	// Logging configuration
	Log LogConfig `yaml:"log" json:"log"`
	
	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`
	
	// SkyWalking configuration
	SkyWalking SkyWalkingConfig `yaml:"skywalking" json:"skywalking"`
}

// ServerConfig defines web server configuration
type ServerConfig struct {
	Address      string        `yaml:"address" json:"address"`
	Port         int           `yaml:"port" json:"port"`
	ReadTimeout  time.Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout" json:"idleTimeout"`
	TLSEnabled   bool          `yaml:"tlsEnabled" json:"tlsEnabled"`
	CertFile     string        `yaml:"certFile" json:"certFile"`
	KeyFile      string        `yaml:"keyFile" json:"keyFile"`
}

// AuthConfig defines authentication configuration
type AuthConfig struct {
	JWTSecret          string        `yaml:"jwtSecret" json:"jwtSecret"`
	JWTExpiration      time.Duration `yaml:"jwtExpiration" json:"jwtExpiration"`
	RefreshExpiration  time.Duration `yaml:"refreshExpiration" json:"refreshExpiration"`
	EnableRateLimit    bool          `yaml:"enableRateLimit" json:"enableRateLimit"`
	RateLimitRequests  int           `yaml:"rateLimitRequests" json:"rateLimitRequests"`
	RateLimitDuration  time.Duration `yaml:"rateLimitDuration" json:"rateLimitDuration"`
	SessionTimeout     time.Duration `yaml:"sessionTimeout" json:"sessionTimeout"`
	MaxConcurrentSessions int        `yaml:"maxConcurrentSessions" json:"maxConcurrentSessions"`
}

// DefaultWebConfig returns default web configuration
func DefaultWebConfig() *WebConfig {
	return &WebConfig{
		APIVersion: "v1",
		Kind:       "WebConfig",
		Server: ServerConfig{
			Address:      "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			TLSEnabled:   false,
		},
		Database: DatabaseConfig{
			Driver:       "mysql",
			Host:         "localhost",
			Port:         3306,
			Database:     "sing_box_web",
			Username:     "root",
			Password:     "",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			MaxLifetime:  time.Hour,
		},
		APIServer: APIServerConnection{
			Address:  "localhost",
			Port:     8081,
			Timeout:  10 * time.Second,
			Insecure: true,
		},
		Auth: AuthConfig{
			JWTSecret:             "default-jwt-secret",
			JWTExpiration:         24 * time.Hour,
			RefreshExpiration:     7 * 24 * time.Hour,
			EnableRateLimit:       true,
			RateLimitRequests:     100,
			RateLimitDuration:     time.Minute,
			SessionTimeout:        30 * time.Minute,
			MaxConcurrentSessions: 5,
		},
		Log: LogConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Address: "0.0.0.0",
			Port:    9090,
			Path:    "/metrics",
		},
		SkyWalking: SkyWalkingConfig{
			Enabled:     false,
			Collector:   "localhost:11800",
			ServiceName: "sing-box-web",
			SampleRate:  1,
		},
	}
}