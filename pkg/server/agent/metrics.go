package agent

import (
	"context"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"

	pbv1 "sing-box-web/pkg/pb/v1"
)

// MetricsCollector collects system metrics
type MetricsCollector struct {
	logger *zap.Logger

	// Current metrics
	metrics   *pbv1.NodeMetrics
	metricsMu sync.RWMutex

	// Shutdown
	shutdownCtx context.Context
	shutdown    context.CancelFunc
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	shutdownCtx, shutdown := context.WithCancel(context.Background())

	return &MetricsCollector{
		logger:      logger.Named("metrics"),
		shutdownCtx: shutdownCtx,
		shutdown:    shutdown,
	}
}

// Start starts the metrics collection
func (m *MetricsCollector) Start(ctx context.Context) error {
	m.logger.Info("starting metrics collection")

	// Start metrics collection loop
	go m.collectLoop()

	return nil
}

// Stop stops the metrics collection
func (m *MetricsCollector) Stop(ctx context.Context) error {
	m.logger.Info("stopping metrics collection")
	m.shutdown()
	return nil
}

// collectLoop collects metrics periodically
func (m *MetricsCollector) collectLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.shutdownCtx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

// collectMetrics collects current system metrics
func (m *MetricsCollector) collectMetrics() {
	metrics := &pbv1.NodeMetrics{
		CpuUsagePercent:       float64(m.getCPUUsage()),
		MemoryUsagePercent:    float64(m.getMemoryUsage()),
		DiskUsagePercent:      float64(m.getDiskUsage()),
		NetworkInBytesPerSec:  m.getNetworkIn(),
		NetworkOutBytesPerSec: m.getNetworkOut(),
		ActiveConnections:     m.getConnections(),
		LoadAverage:           float64(m.getLoadAverage1()),
	}

	m.metricsMu.Lock()
	m.metrics = metrics
	m.metricsMu.Unlock()
}

// GetMetrics returns the current metrics
func (m *MetricsCollector) GetMetrics() *pbv1.NodeMetrics {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()

	if m.metrics == nil {
		return nil
	}

	// Return a copy to avoid race conditions
	return &pbv1.NodeMetrics{
		CpuUsagePercent:       m.metrics.CpuUsagePercent,
		MemoryUsagePercent:    m.metrics.MemoryUsagePercent,
		DiskUsagePercent:      m.metrics.DiskUsagePercent,
		NetworkInBytesPerSec:  m.metrics.NetworkInBytesPerSec,
		NetworkOutBytesPerSec: m.metrics.NetworkOutBytesPerSec,
		ActiveConnections:     m.metrics.ActiveConnections,
		LoadAverage:           m.metrics.LoadAverage,
	}
}

// getCPUUsage gets CPU usage percentage
func (m *MetricsCollector) getCPUUsage() float32 {
	// Simple CPU usage calculation using runtime
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// This is a simplified calculation
	// In a real implementation, you would use system-specific APIs
	return float32(runtime.NumCPU()) * 10.0 // Placeholder
}

// getMemoryUsage gets memory usage percentage
func (m *MetricsCollector) getMemoryUsage() float32 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Calculate memory usage percentage
	// This is simplified - in reality you'd compare against total system memory
	usedMB := float32(memStats.Sys) / 1024 / 1024
	totalMB := float32(8192) // Assume 8GB total memory for demo
	
	return (usedMB / totalMB) * 100
}

// getDiskUsage gets disk usage percentage
func (m *MetricsCollector) getDiskUsage() float32 {
	// Placeholder implementation
	// In a real implementation, you would use syscalls to get disk usage
	return 45.0 // Placeholder: 45% disk usage
}

// getNetworkIn gets network input bytes
func (m *MetricsCollector) getNetworkIn() int64 {
	// Placeholder implementation
	// In a real implementation, you would read from /proc/net/dev or similar
	return 1024 * 1024 * 100 // Placeholder: 100MB
}

// getNetworkOut gets network output bytes
func (m *MetricsCollector) getNetworkOut() int64 {
	// Placeholder implementation
	// In a real implementation, you would read from /proc/net/dev or similar
	return 1024 * 1024 * 80 // Placeholder: 80MB
}

// getLoadAverage1 gets 1-minute load average
func (m *MetricsCollector) getLoadAverage1() float32 {
	// Placeholder implementation
	// In a real implementation, you would read from /proc/loadavg on Linux
	return 0.5 // Placeholder
}

// getLoadAverage5 gets 5-minute load average
func (m *MetricsCollector) getLoadAverage5() float32 {
	// Placeholder implementation
	return 0.7 // Placeholder
}

// getLoadAverage15 gets 15-minute load average
func (m *MetricsCollector) getLoadAverage15() float32 {
	// Placeholder implementation
	return 0.8 // Placeholder
}

// getConnections gets current connection count
func (m *MetricsCollector) getConnections() int32 {
	// Placeholder implementation
	// In a real implementation, you would count active connections
	return 42 // Placeholder
}