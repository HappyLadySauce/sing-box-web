package v1

import "time"

// AgentConfig defines configuration for sing-box-agent service
type AgentConfig struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	
	// Node information
	Node NodeInfo `yaml:"node" json:"node"`
	
	// API server connection
	APIServer APIServerConnection `yaml:"apiServer" json:"apiServer"`
	
	// sing-box configuration
	SingBox SingBoxConfig `yaml:"singBox" json:"singBox"`
	
	// Monitoring configuration
	Monitor MonitorConfig `yaml:"monitor" json:"monitor"`
	
	// Logging configuration
	Log LogConfig `yaml:"log" json:"log"`
	
	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`
	
	// SkyWalking configuration
	SkyWalking SkyWalkingConfig `yaml:"skywalking" json:"skywalking"`
}

// NodeInfo defines node information
type NodeInfo struct {
	NodeID       string            `yaml:"nodeId" json:"nodeId"`
	NodeName     string            `yaml:"nodeName" json:"nodeName"`
	Region       string            `yaml:"region" json:"region"`
	Zone         string            `yaml:"zone" json:"zone"`
	Tags         map[string]string `yaml:"tags" json:"tags"`
	Capabilities []string          `yaml:"capabilities" json:"capabilities"`
	MaxUsers     int               `yaml:"maxUsers" json:"maxUsers"`
}

// SingBoxConfig defines sing-box related configuration
type SingBoxConfig struct {
	BinaryPath     string        `yaml:"binaryPath" json:"binaryPath"`
	ConfigPath     string        `yaml:"configPath" json:"configPath"`
	WorkingDir     string        `yaml:"workingDir" json:"workingDir"`
	LogPath        string        `yaml:"logPath" json:"logPath"`
	RestartDelay   time.Duration `yaml:"restartDelay" json:"restartDelay"`
	HealthCheckURL string        `yaml:"healthCheckUrl" json:"healthCheckUrl"`
	ClashAPI       ClashAPIConfig `yaml:"clashApi" json:"clashApi"`
}

// ClashAPIConfig defines Clash API configuration
type ClashAPIConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Address string `yaml:"address" json:"address"`
	Port    int    `yaml:"port" json:"port"`
	Secret  string `yaml:"secret" json:"secret"`
}

// MonitorConfig defines monitoring configuration
type MonitorConfig struct {
	// Data collection intervals
	SystemMetricsInterval  time.Duration `yaml:"systemMetricsInterval" json:"systemMetricsInterval"`
	TrafficReportInterval  time.Duration `yaml:"trafficReportInterval" json:"trafficReportInterval"`
	HeartbeatInterval      time.Duration `yaml:"heartbeatInterval" json:"heartbeatInterval"`
	
	// Data collection settings
	EnableSystemMetrics    bool `yaml:"enableSystemMetrics" json:"enableSystemMetrics"`
	EnableTrafficReport    bool `yaml:"enableTrafficReport" json:"enableTrafficReport"`
	EnableConnectionStats  bool `yaml:"enableConnectionStats" json:"enableConnectionStats"`
	
	// Local cache settings
	LocalCacheSize         int           `yaml:"localCacheSize" json:"localCacheSize"`
	LocalCacheFlushInterval time.Duration `yaml:"localCacheFlushInterval" json:"localCacheFlushInterval"`
	
	// Retry settings
	MaxRetries     int           `yaml:"maxRetries" json:"maxRetries"`
	RetryBackoff   time.Duration `yaml:"retryBackoff" json:"retryBackoff"`
	RetryTimeout   time.Duration `yaml:"retryTimeout" json:"retryTimeout"`
}

// DefaultAgentConfig returns default agent configuration
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		APIVersion: "v1",
		Kind:       "AgentConfig",
		Node: NodeInfo{
			NodeID:       "node-001",
			NodeName:     "Default Node",
			Region:       "default",
			Zone:         "default",
			Tags:         map[string]string{},
			Capabilities: []string{"user_management", "traffic_stats"},
			MaxUsers:     1000,
		},
		APIServer: APIServerConnection{
			Address:  "localhost",
			Port:     8081,
			Timeout:  10 * time.Second,
			Insecure: true,
		},
		SingBox: SingBoxConfig{
			BinaryPath:     "/usr/local/bin/sing-box",
			ConfigPath:     "/etc/sing-box/config.json",
			WorkingDir:     "/var/lib/sing-box",
			LogPath:        "/var/log/sing-box/sing-box.log",
			RestartDelay:   5 * time.Second,
			HealthCheckURL: "http://127.0.0.1:9090/health",
			ClashAPI: ClashAPIConfig{
				Enabled: true,
				Address: "127.0.0.1",
				Port:    9090,
				Secret:  "",
			},
		},
		Monitor: MonitorConfig{
			SystemMetricsInterval:       30 * time.Second,
			TrafficReportInterval:       5 * time.Minute,
			HeartbeatInterval:          30 * time.Second,
			EnableSystemMetrics:        true,
			EnableTrafficReport:        true,
			EnableConnectionStats:      true,
			LocalCacheSize:             1000,
			LocalCacheFlushInterval:    time.Minute,
			MaxRetries:                 3,
			RetryBackoff:               5 * time.Second,
			RetryTimeout:               30 * time.Second,
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
			Port:    9092,
			Path:    "/metrics",
		},
		SkyWalking: SkyWalkingConfig{
			Enabled:     false,
			Collector:   "localhost:11800",
			ServiceName: "sing-box-agent",
			SampleRate:  1,
		},
	}
}