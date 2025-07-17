package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	grpcclient "sing-box-web/pkg/client/grpc"
	configv1 "sing-box-web/pkg/config/v1"
	pbv1 "sing-box-web/pkg/pb/v1"
)

// ClientManager manages multiple gRPC client connections with load balancing
type ClientManager struct {
	logger *zap.Logger
	config configv1.APIServerConnection

	// Client pool
	clients []*grpcclient.Client
	mutex   sync.RWMutex

	// Load balancing
	currentIndex int
	indexMutex   sync.Mutex

	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClientManager creates a new client manager
func NewClientManager(config configv1.APIServerConnection, logger *zap.Logger) *ClientManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ClientManager{
		logger: logger,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize initializes the client manager with multiple connections
func (cm *ClientManager) Initialize(poolSize int) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if poolSize <= 0 {
		poolSize = 1
	}

	cm.logger.Info("Initializing client pool", zap.Int("pool_size", poolSize))

	for i := 0; i < poolSize; i++ {
		client := grpcclient.NewClient(cm.config, cm.logger.With(zap.Int("client_id", i)))
		if err := client.Connect(); err != nil {
			cm.logger.Error("Failed to connect client", zap.Int("client_id", i), zap.Error(err))
			continue
		}
		cm.clients = append(cm.clients, client)
	}

	if len(cm.clients) == 0 {
		return fmt.Errorf("failed to connect any clients")
	}

	cm.logger.Info("Client pool initialized", zap.Int("connected_clients", len(cm.clients)))
	return nil
}

// GetHealthyClient returns a healthy client using round-robin load balancing
func (cm *ClientManager) GetHealthyClient() *grpcclient.Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if len(cm.clients) == 0 {
		return nil
	}

	// Try to find a healthy client starting from the next index
	startIndex := cm.getNextIndex()
	currentIndex := startIndex

	for {
		client := cm.clients[currentIndex]
		if client.IsConnected() {
			return client
		}

		currentIndex = (currentIndex + 1) % len(cm.clients)
		if currentIndex == startIndex {
			// We've tried all clients
			break
		}
	}

	// No healthy clients found, return the first one (it might recover)
	return cm.clients[0]
}

// getNextIndex returns the next index for load balancing
func (cm *ClientManager) getNextIndex() int {
	cm.indexMutex.Lock()
	defer cm.indexMutex.Unlock()

	cm.currentIndex = (cm.currentIndex + 1) % len(cm.clients)
	return cm.currentIndex
}

// GetManagementClient returns a management service client
func (cm *ClientManager) GetManagementClient() pbv1.ManagementServiceClient {
	client := cm.GetHealthyClient()
	if client == nil {
		return nil
	}
	return client.GetManagementClient()
}

// GetAgentClient returns an agent service client
func (cm *ClientManager) GetAgentClient() pbv1.AgentServiceClient {
	client := cm.GetHealthyClient()
	if client == nil {
		return nil
	}
	return client.GetAgentClient()
}

// CallWithRetry executes a function with retry logic on different clients
func (cm *ClientManager) CallWithRetry(ctx context.Context, fn func(client *grpcclient.Client) error, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		client := cm.GetHealthyClient()
		if client == nil {
			return fmt.Errorf("no healthy clients available")
		}

		err := fn(client)
		if err == nil {
			return nil
		}

		lastErr = err
		cm.logger.Warn("Client call failed", zap.Int("attempt", attempt+1), zap.Error(err))

		if attempt < maxRetries {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second * time.Duration(attempt+1)):
				continue
			}
		}
	}

	return fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

// GetConnectionStats returns connection statistics
func (cm *ClientManager) GetConnectionStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	total := len(cm.clients)
	healthy := 0
	connected := 0

	stats := make([]map[string]interface{}, total)

	for i, client := range cm.clients {
		isConnected := client.IsConnected()
		state := client.GetConnectionState()

		if isConnected {
			connected++
			healthy++
		}

		stats[i] = map[string]interface{}{
			"id":        i,
			"connected": isConnected,
			"state":     state.String(),
		}
	}

	return map[string]interface{}{
		"total":     total,
		"healthy":   healthy,
		"connected": connected,
		"clients":   stats,
	}
}

// Close closes all client connections
func (cm *ClientManager) Close() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.cancel()

	var errors []error
	for i, client := range cm.clients {
		if err := client.Close(); err != nil {
			cm.logger.Error("Failed to close client", zap.Int("client_id", i), zap.Error(err))
			errors = append(errors, err)
		}
	}

	cm.clients = nil

	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d clients", len(errors))
	}

	return nil
}

// Reconnect attempts to reconnect all disconnected clients
func (cm *ClientManager) Reconnect() {
	cm.mutex.RLock()
	clients := make([]*grpcclient.Client, len(cm.clients))
	copy(clients, cm.clients)
	cm.mutex.RUnlock()

	for i, client := range clients {
		if !client.IsConnected() {
			cm.logger.Info("Reconnecting client", zap.Int("client_id", i))
			go func(c *grpcclient.Client, id int) {
				if err := c.Connect(); err != nil {
					cm.logger.Error("Failed to reconnect client", zap.Int("client_id", id), zap.Error(err))
				}
			}(client, i)
		}
	}
}

// StartHealthChecker starts a health checker goroutine
func (cm *ClientManager) StartHealthChecker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-cm.ctx.Done():
				return
			case <-ticker.C:
				cm.checkHealth()
			}
		}
	}()
}

// checkHealth checks the health of all clients
func (cm *ClientManager) checkHealth() {
	stats := cm.GetConnectionStats()
	healthy := stats["healthy"].(int)
	total := stats["total"].(int)

	cm.logger.Debug("Health check", zap.Int("healthy", healthy), zap.Int("total", total))

	if healthy < total/2 {
		cm.logger.Warn("More than half of clients are unhealthy", zap.Int("healthy", healthy), zap.Int("total", total))
		cm.Reconnect()
	}
}

// Global client manager instance
var globalClientManager *ClientManager

// InitGlobalClientManager initializes the global client manager
func InitGlobalClientManager(config configv1.APIServerConnection, logger *zap.Logger, poolSize int) error {
	globalClientManager = NewClientManager(config, logger)
	if err := globalClientManager.Initialize(poolSize); err != nil {
		return err
	}
	globalClientManager.StartHealthChecker(30 * time.Second)
	return nil
}

// GetGlobalClientManager returns the global client manager
func GetGlobalClientManager() *ClientManager {
	return globalClientManager
}

// GetGlobalManagementClient returns a management client from the global manager
func GetGlobalManagementClient() pbv1.ManagementServiceClient {
	if globalClientManager == nil {
		return nil
	}
	return globalClientManager.GetManagementClient()
}

// GetGlobalAgentClient returns an agent client from the global manager
func GetGlobalAgentClient() pbv1.AgentServiceClient {
	if globalClientManager == nil {
		return nil
	}
	return globalClientManager.GetAgentClient()
}
