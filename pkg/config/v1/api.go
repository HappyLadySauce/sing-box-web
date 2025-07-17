package v1

import "time"

// APIConfig defines configuration for sing-box-api service
type APIConfig struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	
	// gRPC server configuration
	GRPC GRPCServerConfig `yaml:"grpc" json:"grpc"`
	
	// Database configuration
	Database DatabaseConfig `yaml:"database" json:"database"`
	
	// Logging configuration
	Log LogConfig `yaml:"log" json:"log"`
	
	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`
	
	// SkyWalking configuration
	SkyWalking SkyWalkingConfig `yaml:"skywalking" json:"skywalking"`
	
	// Business configuration
	Business BusinessConfig `yaml:"business" json:"business"`
}

// GRPCServerConfig defines gRPC server configuration
type GRPCServerConfig struct {
	Address           string        `yaml:"address" json:"address"`
	Port              int           `yaml:"port" json:"port"`
	MaxRecvMsgSize    int           `yaml:"maxRecvMsgSize" json:"maxRecvMsgSize"`
	MaxSendMsgSize    int           `yaml:"maxSendMsgSize" json:"maxSendMsgSize"`
	ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`
	KeepaliveTime     time.Duration `yaml:"keepaliveTime" json:"keepaliveTime"`
	KeepaliveTimeout  time.Duration `yaml:"keepaliveTimeout" json:"keepaliveTimeout"`
	TLSEnabled        bool          `yaml:"tlsEnabled" json:"tlsEnabled"`
	CertFile          string        `yaml:"certFile" json:"certFile"`
	KeyFile           string        `yaml:"keyFile" json:"keyFile"`
	ClientCAs         string        `yaml:"clientCAs" json:"clientCAs"`
}

// BusinessConfig defines business logic configuration
type BusinessConfig struct {
	// Traffic management
	Traffic TrafficConfig `yaml:"traffic" json:"traffic"`
	
	// Node management
	Node NodeConfig `yaml:"node" json:"node"`
	
	// User management
	User UserConfig `yaml:"user" json:"user"`
	
	// Alert configuration
	Alert AlertConfig `yaml:"alert" json:"alert"`
}

// TrafficConfig defines traffic management configuration
type TrafficConfig struct {
	ReportInterval     time.Duration `yaml:"reportInterval" json:"reportInterval"`
	BatchSize          int           `yaml:"batchSize" json:"batchSize"`
	RetentionDays      int           `yaml:"retentionDays" json:"retentionDays"`
	EnableCompression  bool          `yaml:"enableCompression" json:"enableCompression"`
	EnableAggregation  bool          `yaml:"enableAggregation" json:"enableAggregation"`
	AggregationWindow  time.Duration `yaml:"aggregationWindow" json:"aggregationWindow"`
}

// NodeConfig defines node management configuration
type NodeConfig struct {
	HeartbeatInterval  time.Duration `yaml:"heartbeatInterval" json:"heartbeatInterval"`
	HeartbeatTimeout   time.Duration `yaml:"heartbeatTimeout" json:"heartbeatTimeout"`
	MaxOfflineTime     time.Duration `yaml:"maxOfflineTime" json:"maxOfflineTime"`
	ConfigSyncInterval time.Duration `yaml:"configSyncInterval" json:"configSyncInterval"`
	MaxRetries         int           `yaml:"maxRetries" json:"maxRetries"`
	RetryBackoff       time.Duration `yaml:"retryBackoff" json:"retryBackoff"`
}

// UserConfig defines user management configuration
type UserConfig struct {
	MaxUsersPerNode    int           `yaml:"maxUsersPerNode" json:"maxUsersPerNode"`
	DefaultPlanID      int64         `yaml:"defaultPlanID" json:"defaultPlanID"`
	PasswordMinLength  int           `yaml:"passwordMinLength" json:"passwordMinLength"`
	EnableUserLimit    bool          `yaml:"enableUserLimit" json:"enableUserLimit"`
	UserLimitCheckInterval time.Duration `yaml:"userLimitCheckInterval" json:"userLimitCheckInterval"`
}

// AlertConfig defines alert configuration
type AlertConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	SMTPHost          string        `yaml:"smtpHost" json:"smtpHost"`
	SMTPPort          int           `yaml:"smtpPort" json:"smtpPort"`
	SMTPUser          string        `yaml:"smtpUser" json:"smtpUser"`
	SMTPPassword      string        `yaml:"smtpPassword" json:"smtpPassword"`
	DefaultRecipients []string      `yaml:"defaultRecipients" json:"defaultRecipients"`
	AlertCooldown     time.Duration `yaml:"alertCooldown" json:"alertCooldown"`
}

// DefaultAPIConfig returns default API configuration
func DefaultAPIConfig() *APIConfig {
	return &APIConfig{
		APIVersion: "v1",
		Kind:       "APIConfig",
		GRPC: GRPCServerConfig{
			Address:           "0.0.0.0",
			Port:              8081,
			MaxRecvMsgSize:    4 * 1024 * 1024, // 4MB
			MaxSendMsgSize:    4 * 1024 * 1024, // 4MB
			ConnectionTimeout: 10 * time.Second,
			KeepaliveTime:     30 * time.Second,
			KeepaliveTimeout:  5 * time.Second,
			TLSEnabled:        false,
		},
		Database: DatabaseConfig{
			Driver:       "mysql",
			Host:         "localhost",
			Port:         3306,
			Database:     "sing_box_api",
			Username:     "root",
			Password:     "",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			MaxLifetime:  time.Hour,
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
			Port:    9091,
			Path:    "/metrics",
		},
		SkyWalking: SkyWalkingConfig{
			Enabled:     false,
			Collector:   "localhost:11800",
			ServiceName: "sing-box-api",
			SampleRate:  1,
		},
		Business: BusinessConfig{
			Traffic: TrafficConfig{
				ReportInterval:    5 * time.Minute,
				BatchSize:         1000,
				RetentionDays:     90,
				EnableCompression: true,
				EnableAggregation: true,
				AggregationWindow: time.Hour,
			},
			Node: NodeConfig{
				HeartbeatInterval:  30 * time.Second,
				HeartbeatTimeout:   10 * time.Second,
				MaxOfflineTime:     5 * time.Minute,
				ConfigSyncInterval: 10 * time.Minute,
				MaxRetries:         3,
				RetryBackoff:       5 * time.Second,
			},
			User: UserConfig{
				MaxUsersPerNode:        1000,
				DefaultPlanID:          1,
				PasswordMinLength:      8,
				EnableUserLimit:        true,
				UserLimitCheckInterval: time.Hour,
			},
			Alert: AlertConfig{
				Enabled:       false,
				SMTPHost:      "localhost",
				SMTPPort:      587,
				AlertCooldown: 15 * time.Minute,
			},
		},
	}
}