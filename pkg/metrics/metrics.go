package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	configv1 "sing-box-web/pkg/config/v1"
)

// MetricsCollector manages Prometheus metrics collection
type MetricsCollector struct {
	registry *prometheus.Registry
	logger   *zap.Logger

	// HTTP metrics
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpActiveConnections prometheus.Gauge

	// gRPC metrics
	grpcRequestsTotal   *prometheus.CounterVec
	grpcRequestDuration *prometheus.HistogramVec

	// Node metrics
	nodeStatus      *prometheus.GaugeVec
	nodeLastSeen    *prometheus.GaugeVec
	nodeUserCount   *prometheus.GaugeVec
	nodeConnections *prometheus.GaugeVec

	// User metrics
	userTotal        prometheus.Gauge
	userActiveTotal  prometheus.Gauge
	userTrafficBytes *prometheus.CounterVec

	// System metrics
	systemUptime      prometheus.Gauge
	systemMemoryUsage prometheus.Gauge
	systemCPUUsage    prometheus.Gauge
	systemGoroutines  prometheus.Gauge

	// Database metrics
	dbConnections   *prometheus.GaugeVec
	dbQueryDuration *prometheus.HistogramVec
	dbQueryTotal    *prometheus.CounterVec

	// Business metrics
	trafficTotalBytes     *prometheus.CounterVec
	traffic24hBytes       *prometheus.GaugeVec
	userQuotaUsagePercent *prometheus.GaugeVec
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	registry := prometheus.NewRegistry()

	c := &MetricsCollector{
		registry: registry,
		logger:   logger,
	}

	c.initMetrics()
	c.registerMetrics()

	return c
}

// initMetrics initializes all Prometheus metrics
func (c *MetricsCollector) initMetrics() {
	// HTTP metrics
	c.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sing_box_web_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	c.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sing_box_web_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	c.httpActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sing_box_web_http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	// gRPC metrics
	c.grpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sing_box_api_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	c.grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sing_box_api_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	// Node metrics
	c.nodeStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sing_box_node_status",
			Help: "Node status (1=online, 0=offline)",
		},
		[]string{"node_id", "node_name"},
	)

	c.nodeLastSeen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sing_box_node_last_seen_timestamp",
			Help: "Timestamp of node last seen",
		},
		[]string{"node_id", "node_name"},
	)

	c.nodeUserCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sing_box_node_user_count",
			Help: "Number of users on each node",
		},
		[]string{"node_id", "node_name"},
	)

	c.nodeConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sing_box_node_connections",
			Help: "Number of active connections on each node",
		},
		[]string{"node_id", "node_name"},
	)

	// User metrics
	c.userTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sing_box_users_total",
			Help: "Total number of users",
		},
	)

	c.userActiveTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sing_box_users_active_total",
			Help: "Number of active users",
		},
	)

	c.userTrafficBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sing_box_user_traffic_bytes_total",
			Help: "Total user traffic in bytes",
		},
		[]string{"user_id", "direction", "node_id"},
	)

	// System metrics
	c.systemUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sing_box_system_uptime_seconds",
			Help: "System uptime in seconds",
		},
	)

	c.systemMemoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sing_box_system_memory_usage_bytes",
			Help: "System memory usage in bytes",
		},
	)

	c.systemCPUUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sing_box_system_cpu_usage_percent",
			Help: "System CPU usage percentage",
		},
	)

	c.systemGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sing_box_system_goroutines",
			Help: "Number of goroutines",
		},
	)

	// Database metrics
	c.dbConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sing_box_db_connections",
			Help: "Number of database connections",
		},
		[]string{"state"}, // open, idle, in_use
	)

	c.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sing_box_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"operation"},
	)

	c.dbQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sing_box_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"},
	)

	// Business metrics
	c.trafficTotalBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sing_box_traffic_total_bytes",
			Help: "Total traffic in bytes",
		},
		[]string{"direction", "node_id"},
	)

	c.traffic24hBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sing_box_traffic_24h_bytes",
			Help: "Traffic in last 24 hours in bytes",
		},
		[]string{"direction", "node_id"},
	)

	c.userQuotaUsagePercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sing_box_user_quota_usage_percent",
			Help: "User quota usage percentage",
		},
		[]string{"user_id", "node_id"},
	)
}

// registerMetrics registers all metrics with the registry
func (c *MetricsCollector) registerMetrics() {
	// HTTP metrics
	c.registry.MustRegister(c.httpRequestsTotal)
	c.registry.MustRegister(c.httpRequestDuration)
	c.registry.MustRegister(c.httpActiveConnections)

	// gRPC metrics
	c.registry.MustRegister(c.grpcRequestsTotal)
	c.registry.MustRegister(c.grpcRequestDuration)

	// Node metrics
	c.registry.MustRegister(c.nodeStatus)
	c.registry.MustRegister(c.nodeLastSeen)
	c.registry.MustRegister(c.nodeUserCount)
	c.registry.MustRegister(c.nodeConnections)

	// User metrics
	c.registry.MustRegister(c.userTotal)
	c.registry.MustRegister(c.userActiveTotal)
	c.registry.MustRegister(c.userTrafficBytes)

	// System metrics
	c.registry.MustRegister(c.systemUptime)
	c.registry.MustRegister(c.systemMemoryUsage)
	c.registry.MustRegister(c.systemCPUUsage)
	c.registry.MustRegister(c.systemGoroutines)

	// Database metrics
	c.registry.MustRegister(c.dbConnections)
	c.registry.MustRegister(c.dbQueryDuration)
	c.registry.MustRegister(c.dbQueryTotal)

	// Business metrics
	c.registry.MustRegister(c.trafficTotalBytes)
	c.registry.MustRegister(c.traffic24hBytes)
	c.registry.MustRegister(c.userQuotaUsagePercent)

	// Add Go runtime metrics
	c.registry.MustRegister(prometheus.NewGoCollector())
	c.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

// GetHandler returns the Prometheus HTTP handler
func (c *MetricsCollector) GetHandler() http.Handler {
	return promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{})
}

// HTTP Metrics

// RecordHTTPRequest records an HTTP request
func (c *MetricsCollector) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	c.httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
	c.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// SetHTTPActiveConnections sets the number of active HTTP connections
func (c *MetricsCollector) SetHTTPActiveConnections(count float64) {
	c.httpActiveConnections.Set(count)
}

// gRPC Metrics

// RecordGRPCRequest records a gRPC request
func (c *MetricsCollector) RecordGRPCRequest(service, method, status string, duration time.Duration) {
	c.grpcRequestsTotal.WithLabelValues(service, method, status).Inc()
	c.grpcRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// Node Metrics

// SetNodeStatus sets the status of a node
func (c *MetricsCollector) SetNodeStatus(nodeID, nodeName string, online bool) {
	status := 0.0
	if online {
		status = 1.0
	}
	c.nodeStatus.WithLabelValues(nodeID, nodeName).Set(status)
}

// SetNodeLastSeen sets the last seen timestamp of a node
func (c *MetricsCollector) SetNodeLastSeen(nodeID, nodeName string, timestamp time.Time) {
	c.nodeLastSeen.WithLabelValues(nodeID, nodeName).Set(float64(timestamp.Unix()))
}

// SetNodeUserCount sets the user count for a node
func (c *MetricsCollector) SetNodeUserCount(nodeID, nodeName string, count int) {
	c.nodeUserCount.WithLabelValues(nodeID, nodeName).Set(float64(count))
}

// SetNodeConnections sets the connection count for a node
func (c *MetricsCollector) SetNodeConnections(nodeID, nodeName string, count int) {
	c.nodeConnections.WithLabelValues(nodeID, nodeName).Set(float64(count))
}

// User Metrics

// SetUserTotal sets the total number of users
func (c *MetricsCollector) SetUserTotal(count float64) {
	c.userTotal.Set(count)
}

// SetUserActiveTotal sets the number of active users
func (c *MetricsCollector) SetUserActiveTotal(count float64) {
	c.userActiveTotal.Set(count)
}

// RecordUserTraffic records user traffic
func (c *MetricsCollector) RecordUserTraffic(userID, direction, nodeID string, bytes int64) {
	c.userTrafficBytes.WithLabelValues(userID, direction, nodeID).Add(float64(bytes))
}

// System Metrics

// SetSystemUptime sets the system uptime
func (c *MetricsCollector) SetSystemUptime(seconds float64) {
	c.systemUptime.Set(seconds)
}

// SetSystemMemoryUsage sets the system memory usage
func (c *MetricsCollector) SetSystemMemoryUsage(bytes float64) {
	c.systemMemoryUsage.Set(bytes)
}

// SetSystemCPUUsage sets the system CPU usage
func (c *MetricsCollector) SetSystemCPUUsage(percent float64) {
	c.systemCPUUsage.Set(percent)
}

// SetSystemGoroutines sets the number of goroutines
func (c *MetricsCollector) SetSystemGoroutines(count float64) {
	c.systemGoroutines.Set(count)
}

// Database Metrics

// SetDBConnections sets database connection counts
func (c *MetricsCollector) SetDBConnections(open, idle, inUse int) {
	c.dbConnections.WithLabelValues("open").Set(float64(open))
	c.dbConnections.WithLabelValues("idle").Set(float64(idle))
	c.dbConnections.WithLabelValues("in_use").Set(float64(inUse))
}

// RecordDBQuery records a database query
func (c *MetricsCollector) RecordDBQuery(operation, status string, duration time.Duration) {
	c.dbQueryTotal.WithLabelValues(operation, status).Inc()
	c.dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// Business Metrics

// RecordTraffic records total traffic
func (c *MetricsCollector) RecordTraffic(direction, nodeID string, bytes int64) {
	c.trafficTotalBytes.WithLabelValues(direction, nodeID).Add(float64(bytes))
}

// SetTraffic24h sets 24-hour traffic
func (c *MetricsCollector) SetTraffic24h(direction, nodeID string, bytes int64) {
	c.traffic24hBytes.WithLabelValues(direction, nodeID).Set(float64(bytes))
}

// SetUserQuotaUsage sets user quota usage percentage
func (c *MetricsCollector) SetUserQuotaUsage(userID, nodeID string, percent float64) {
	c.userQuotaUsagePercent.WithLabelValues(userID, nodeID).Set(percent)
}

// StartMetricsServer starts the metrics HTTP server
func (c *MetricsCollector) StartMetricsServer(config configv1.MetricsConfig) error {
	if !config.Enabled {
		c.logger.Info("Metrics server disabled")
		return nil
	}

	addr := config.Address + ":" + strconv.Itoa(config.Port)
	c.logger.Info("Starting metrics server", zap.String("address", addr), zap.String("path", config.Path))

	mux := http.NewServeMux()
	mux.Handle(config.Path, c.GetHandler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.logger.Error("Metrics server failed", zap.Error(err))
		}
	}()

	return nil
}

// Global metrics instance
var globalMetrics *MetricsCollector

// InitGlobalMetrics initializes the global metrics collector
func InitGlobalMetrics(logger *zap.Logger) {
	globalMetrics = NewMetricsCollector(logger)
}

// GetGlobalMetrics returns the global metrics collector
func GetGlobalMetrics() *MetricsCollector {
	return globalMetrics
}

// Convenience functions for global metrics

// RecordHTTPRequest records an HTTP request using global metrics
func RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	if globalMetrics != nil {
		globalMetrics.RecordHTTPRequest(method, path, statusCode, duration)
	}
}

// RecordGRPCRequest records a gRPC request using global metrics
func RecordGRPCRequest(service, method, status string, duration time.Duration) {
	if globalMetrics != nil {
		globalMetrics.RecordGRPCRequest(service, method, status, duration)
	}
}

// SetNodeStatus sets node status using global metrics
func SetNodeStatus(nodeID, nodeName string, online bool) {
	if globalMetrics != nil {
		globalMetrics.SetNodeStatus(nodeID, nodeName, online)
	}
}

// RecordUserTraffic records user traffic using global metrics
func RecordUserTraffic(userID, direction, nodeID string, bytes int64) {
	if globalMetrics != nil {
		globalMetrics.RecordUserTraffic(userID, direction, nodeID, bytes)
	}
}

// RecordDBQuery records a database query using global metrics
func RecordDBQuery(operation, status string, duration time.Duration) {
	if globalMetrics != nil {
		globalMetrics.RecordDBQuery(operation, status, duration)
	}
}
