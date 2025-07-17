package validation

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	configv1 "sing-box-web/pkg/config/v1"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%v': %s", e.Field, e.Value, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator provides configuration validation functions
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// addError adds a validation error
func (v *Validator) addError(field string, value interface{}, message string) {
	v.errors = append(v.errors, &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	})
}

// Validate runs all validation checks and returns errors if any
func (v *Validator) Validate() error {
	if len(v.errors) > 0 {
		return v.errors
	}
	return nil
}

// ValidateWebConfig validates web service configuration
func ValidateWebConfig(config *configv1.WebConfig) error {
	validator := NewValidator()

	// Validate API version and kind
	validator.validateAPIVersion(config.APIVersion)
	validator.validateKind(config.Kind, "WebConfig")

	// Validate server configuration
	validator.validateServerConfig(config.Server)

	// Validate database configuration
	validator.validateDatabaseConfig(config.Database)

	// Validate API server connection
	validator.validateAPIServerConnection(config.APIServer)

	// Validate auth configuration
	validator.validateAuthConfig(config.Auth)

	// Validate log configuration
	validator.validateLogConfig(config.Log)

	// Validate metrics configuration
	validator.validateMetricsConfig(config.Metrics)

	// Validate SkyWalking configuration
	validator.validateSkyWalkingConfig(config.SkyWalking)

	return validator.Validate()
}

// ValidateAPIConfig validates API service configuration
func ValidateAPIConfig(config *configv1.APIConfig) error {
	validator := NewValidator()

	// Validate API version and kind
	validator.validateAPIVersion(config.APIVersion)
	validator.validateKind(config.Kind, "APIConfig")

	// Validate gRPC server configuration
	validator.validateGRPCServerConfig(config.GRPC)

	// Validate database configuration
	validator.validateDatabaseConfig(config.Database)

	// Validate log configuration
	validator.validateLogConfig(config.Log)

	// Validate metrics configuration
	validator.validateMetricsConfig(config.Metrics)

	// Validate SkyWalking configuration
	validator.validateSkyWalkingConfig(config.SkyWalking)

	// Validate business configuration
	validator.validateBusinessConfig(config.Business)

	return validator.Validate()
}

// ValidateAgentConfig validates agent service configuration
func ValidateAgentConfig(config *configv1.AgentConfig) error {
	validator := NewValidator()

	// Validate API version and kind
	validator.validateAPIVersion(config.APIVersion)
	validator.validateKind(config.Kind, "AgentConfig")

	// Validate node information
	validator.validateNodeInfo(config.Node)

	// Validate API server connection
	validator.validateAPIServerConnection(config.APIServer)

	// Validate sing-box configuration
	validator.validateSingBoxConfig(config.SingBox)

	// Validate monitor configuration
	validator.validateMonitorConfig(config.Monitor)

	// Validate log configuration
	validator.validateLogConfig(config.Log)

	// Validate metrics configuration
	validator.validateMetricsConfig(config.Metrics)

	// Validate SkyWalking configuration
	validator.validateSkyWalkingConfig(config.SkyWalking)

	return validator.Validate()
}

// Common validation functions

func (v *Validator) validateAPIVersion(version string) {
	if version == "" {
		v.addError("apiVersion", version, "apiVersion cannot be empty")
	} else if !regexp.MustCompile(`^v\d+$`).MatchString(version) {
		v.addError("apiVersion", version, "apiVersion must be in format 'v1', 'v2', etc.")
	}
}

func (v *Validator) validateKind(kind, expected string) {
	if kind == "" {
		v.addError("kind", kind, "kind cannot be empty")
	} else if kind != expected {
		v.addError("kind", kind, fmt.Sprintf("kind must be '%s'", expected))
	}
}

func (v *Validator) validateServerConfig(config configv1.ServerConfig) {
	v.validateAddress(config.Address, "server.address")
	v.validatePort(config.Port, "server.port")
	v.validateDuration(config.ReadTimeout, "server.readTimeout")
	v.validateDuration(config.WriteTimeout, "server.writeTimeout")
	v.validateDuration(config.IdleTimeout, "server.idleTimeout")

	if config.TLSEnabled {
		v.validateFilePath(config.CertFile, "server.certFile")
		v.validateFilePath(config.KeyFile, "server.keyFile")
	}
}

func (v *Validator) validateDatabaseConfig(config configv1.DatabaseConfig) {
	if config.Driver == "" {
		v.addError("database.driver", config.Driver, "database driver cannot be empty")
	} else if config.Driver != "mysql" && config.Driver != "postgres" && config.Driver != "sqlite" {
		v.addError("database.driver", config.Driver, "database driver must be 'mysql', 'postgres', or 'sqlite'")
	}

	// For SQLite, host and port are not required
	if config.Driver != "sqlite" {
		v.validateAddress(config.Host, "database.host")
		v.validatePort(config.Port, "database.port")

		if config.Username == "" {
			v.addError("database.username", config.Username, "database username cannot be empty")
		}
	}

	if config.Database == "" {
		v.addError("database.database", config.Database, "database name cannot be empty")
	}

	if config.MaxIdleConns <= 0 {
		v.addError("database.maxIdleConns", config.MaxIdleConns, "maxIdleConns must be greater than 0")
	}

	if config.MaxOpenConns <= 0 {
		v.addError("database.maxOpenConns", config.MaxOpenConns, "maxOpenConns must be greater than 0")
	}

	if config.MaxIdleConns > config.MaxOpenConns {
		v.addError("database.maxIdleConns", config.MaxIdleConns, "maxIdleConns cannot be greater than maxOpenConns")
	}
}

func (v *Validator) validateAPIServerConnection(config configv1.APIServerConnection) {
	v.validateAddress(config.Address, "apiServer.address")
	v.validatePort(config.Port, "apiServer.port")
	v.validateDuration(config.Timeout, "apiServer.timeout")

	if !config.Insecure {
		v.validateFilePath(config.CertFile, "apiServer.certFile")
		v.validateFilePath(config.KeyFile, "apiServer.keyFile")
		v.validateFilePath(config.CAFile, "apiServer.caFile")
	}
}

func (v *Validator) validateAuthConfig(config configv1.AuthConfig) {
	if config.JWTSecret == "" {
		v.addError("auth.jwtSecret", config.JWTSecret, "JWT secret cannot be empty")
	} else if len(config.JWTSecret) < 32 {
		v.addError("auth.jwtSecret", config.JWTSecret, "JWT secret must be at least 32 characters long")
	}

	v.validateDuration(config.JWTExpiration, "auth.jwtExpiration")
	v.validateDuration(config.RefreshExpiration, "auth.refreshExpiration")
	v.validateDuration(config.SessionTimeout, "auth.sessionTimeout")

	if config.RateLimitRequests <= 0 {
		v.addError("auth.rateLimitRequests", config.RateLimitRequests, "rateLimitRequests must be greater than 0")
	}

	if config.MaxConcurrentSessions <= 0 {
		v.addError("auth.maxConcurrentSessions", config.MaxConcurrentSessions, "maxConcurrentSessions must be greater than 0")
	}
}

func (v *Validator) validateLogConfig(config configv1.LogConfig) {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	if !contains(validLevels, config.Level) {
		v.addError("log.level", config.Level, fmt.Sprintf("log level must be one of: %s", strings.Join(validLevels, ", ")))
	}

	validFormats := []string{"json", "text"}
	if !contains(validFormats, config.Format) {
		v.addError("log.format", config.Format, fmt.Sprintf("log format must be one of: %s", strings.Join(validFormats, ", ")))
	}

	if config.Output != "stdout" && config.Output != "stderr" && !filepath.IsAbs(config.Output) {
		v.addError("log.output", config.Output, "log output must be 'stdout', 'stderr', or an absolute file path")
	}
}

func (v *Validator) validateMetricsConfig(config configv1.MetricsConfig) {
	if config.Enabled {
		v.validateAddress(config.Address, "metrics.address")
		v.validatePort(config.Port, "metrics.port")

		if config.Path == "" {
			v.addError("metrics.path", config.Path, "metrics path cannot be empty")
		} else if !strings.HasPrefix(config.Path, "/") {
			v.addError("metrics.path", config.Path, "metrics path must start with '/'")
		}
	}
}

func (v *Validator) validateSkyWalkingConfig(config configv1.SkyWalkingConfig) {
	if config.Enabled {
		if config.Collector == "" {
			v.addError("skywalking.collector", config.Collector, "SkyWalking collector cannot be empty")
		} else {
			v.validateURL(config.Collector, "skywalking.collector")
		}

		if config.ServiceName == "" {
			v.addError("skywalking.serviceName", config.ServiceName, "SkyWalking service name cannot be empty")
		}

		if config.SampleRate < 0 || config.SampleRate > 10000 {
			v.addError("skywalking.sampleRate", config.SampleRate, "SkyWalking sample rate must be between 0 and 10000")
		}
	}
}

func (v *Validator) validateGRPCServerConfig(config configv1.GRPCServerConfig) {
	v.validateAddress(config.Address, "grpc.address")
	v.validatePort(config.Port, "grpc.port")
	v.validateDuration(config.ConnectionTimeout, "grpc.connectionTimeout")
	v.validateDuration(config.KeepaliveTime, "grpc.keepaliveTime")
	v.validateDuration(config.KeepaliveTimeout, "grpc.keepaliveTimeout")

	if config.MaxRecvMsgSize <= 0 {
		v.addError("grpc.maxRecvMsgSize", config.MaxRecvMsgSize, "maxRecvMsgSize must be greater than 0")
	}

	if config.MaxSendMsgSize <= 0 {
		v.addError("grpc.maxSendMsgSize", config.MaxSendMsgSize, "maxSendMsgSize must be greater than 0")
	}

	if config.TLSEnabled {
		v.validateFilePath(config.CertFile, "grpc.certFile")
		v.validateFilePath(config.KeyFile, "grpc.keyFile")
		if config.ClientCAs != "" {
			v.validateFilePath(config.ClientCAs, "grpc.clientCAs")
		}
	}
}

func (v *Validator) validateBusinessConfig(config configv1.BusinessConfig) {
	// Validate traffic config
	v.validateDuration(config.Traffic.ReportInterval, "business.traffic.reportInterval")
	if config.Traffic.BatchSize <= 0 {
		v.addError("business.traffic.batchSize", config.Traffic.BatchSize, "batch size must be greater than 0")
	}
	if config.Traffic.RetentionDays <= 0 {
		v.addError("business.traffic.retentionDays", config.Traffic.RetentionDays, "retention days must be greater than 0")
	}

	// Validate node config
	v.validateDuration(config.Node.HeartbeatInterval, "business.node.heartbeatInterval")
	v.validateDuration(config.Node.HeartbeatTimeout, "business.node.heartbeatTimeout")
	v.validateDuration(config.Node.MaxOfflineTime, "business.node.maxOfflineTime")
	v.validateDuration(config.Node.ConfigSyncInterval, "business.node.configSyncInterval")

	// Validate user config
	if config.User.MaxUsersPerNode <= 0 {
		v.addError("business.user.maxUsersPerNode", config.User.MaxUsersPerNode, "max users per node must be greater than 0")
	}
	if config.User.PasswordMinLength < 6 {
		v.addError("business.user.passwordMinLength", config.User.PasswordMinLength, "password min length must be at least 6")
	}
}

func (v *Validator) validateNodeInfo(config configv1.NodeInfo) {
	if config.NodeID == "" {
		v.addError("node.nodeId", config.NodeID, "node ID cannot be empty")
	}

	if config.NodeName == "" {
		v.addError("node.nodeName", config.NodeName, "node name cannot be empty")
	}

	if config.MaxUsers <= 0 {
		v.addError("node.maxUsers", config.MaxUsers, "max users must be greater than 0")
	}
}

func (v *Validator) validateSingBoxConfig(config configv1.SingBoxConfig) {
	v.validateFilePath(config.BinaryPath, "singBox.binaryPath")
	v.validateFilePath(config.ConfigPath, "singBox.configPath")
	v.validateDuration(config.RestartDelay, "singBox.restartDelay")

	if config.ClashAPI.Enabled {
		v.validateAddress(config.ClashAPI.Address, "singBox.clashApi.address")
		v.validatePort(config.ClashAPI.Port, "singBox.clashApi.port")
	}
}

func (v *Validator) validateMonitorConfig(config configv1.MonitorConfig) {
	v.validateDuration(config.SystemMetricsInterval, "monitor.systemMetricsInterval")
	v.validateDuration(config.TrafficReportInterval, "monitor.trafficReportInterval")
	v.validateDuration(config.HeartbeatInterval, "monitor.heartbeatInterval")
	v.validateDuration(config.LocalCacheFlushInterval, "monitor.localCacheFlushInterval")
	v.validateDuration(config.RetryBackoff, "monitor.retryBackoff")
	v.validateDuration(config.RetryTimeout, "monitor.retryTimeout")

	if config.LocalCacheSize <= 0 {
		v.addError("monitor.localCacheSize", config.LocalCacheSize, "local cache size must be greater than 0")
	}

	if config.MaxRetries <= 0 {
		v.addError("monitor.maxRetries", config.MaxRetries, "max retries must be greater than 0")
	}
}

// Helper validation functions

func (v *Validator) validateAddress(address, field string) {
	if address == "" {
		v.addError(field, address, "address cannot be empty")
		return
	}

	if address != "0.0.0.0" && address != "localhost" && net.ParseIP(address) == nil {
		v.addError(field, address, "address must be a valid IP address, 'localhost', or '0.0.0.0'")
	}
}

func (v *Validator) validatePort(port int, field string) {
	if port <= 0 || port > 65535 {
		v.addError(field, port, "port must be between 1 and 65535")
	}
}

func (v *Validator) validateDuration(duration time.Duration, field string) {
	if duration <= 0 {
		v.addError(field, duration, "duration must be greater than 0")
	}
}

func (v *Validator) validateFilePath(path, field string) {
	if path == "" {
		v.addError(field, path, "file path cannot be empty")
		return
	}

	if !filepath.IsAbs(path) {
		v.addError(field, path, "file path must be absolute")
		return
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		v.addError(field, path, "file does not exist")
	}
}

func (v *Validator) validateURL(urlStr, field string) {
	if urlStr == "" {
		v.addError(field, urlStr, "URL cannot be empty")
		return
	}

	// Try to parse as URL first
	if _, err := url.Parse(urlStr); err != nil {
		// If URL parsing fails, try to parse as host:port
		if _, err := net.ResolveTCPAddr("tcp", urlStr); err != nil {
			v.addError(field, urlStr, "must be a valid URL or host:port")
		}
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return p > 0 && p <= 65535
}
