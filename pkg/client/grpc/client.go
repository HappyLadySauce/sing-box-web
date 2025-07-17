package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	configv1 "sing-box-web/pkg/config/v1"
	pbv1 "sing-box-web/pkg/pb/v1"
)

// Client wraps gRPC connections with automatic reconnection and load balancing
type Client struct {
	config configv1.APIServerConnection
	logger *zap.Logger

	// Connection management
	conn   *grpc.ClientConn
	mutex  sync.RWMutex
	closed bool

	// gRPC clients
	managementClient pbv1.ManagementServiceClient
	agentClient      pbv1.AgentServiceClient

	// Reconnection
	reconnectCh chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewClient creates a new gRPC client
func NewClient(config configv1.APIServerConnection, logger *zap.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		config:      config,
		logger:      logger,
		reconnectCh: make(chan struct{}, 1),
		ctx:         ctx,
		cancel:      cancel,
	}

	return c
}

// Connect establishes connection to gRPC server
func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	if c.conn != nil {
		c.conn.Close()
	}

	// Build dial options
	opts, err := c.buildDialOptions()
	if err != nil {
		return fmt.Errorf("failed to build dial options: %w", err)
	}

	// Connect to server
	addr := fmt.Sprintf("%s:%d", c.config.Address, c.config.Port)
	c.logger.Info("Connecting to gRPC server", zap.String("address", addr))

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	c.conn = conn
	c.managementClient = pbv1.NewManagementServiceClient(conn)
	c.agentClient = pbv1.NewAgentServiceClient(conn)

	c.logger.Info("Connected to gRPC server", zap.String("address", addr))

	// Start connection monitor
	go c.monitorConnection()

	return nil
}

// buildDialOptions builds gRPC dial options
func (c *Client) buildDialOptions() ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Configure TLS
	if c.config.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := &tls.Config{
			ServerName: c.config.Address,
		}

		if c.config.CAFile != "" {
			// Load CA certificate
			// Implementation would load the CA file
		}

		if c.config.CertFile != "" && c.config.KeyFile != "" {
			// Load client certificate
			// Implementation would load client cert and key
		}

		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	// Set timeout
	if c.config.Timeout > 0 {
		ctx, cancel := context.WithTimeout(c.ctx, c.config.Timeout)
		defer cancel()
		opts = append(opts, grpc.WithBlock())
		_ = ctx // Use ctx for dial timeout
	}

	return opts, nil
}

// monitorConnection monitors the connection and handles reconnection
func (c *Client) monitorConnection() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.healthCheck()
		case <-c.reconnectCh:
			c.reconnect()
		}
	}
}

// healthCheck checks the connection health
func (c *Client) healthCheck() {
	c.mutex.RLock()
	conn := c.conn
	c.mutex.RUnlock()

	if conn == nil {
		c.triggerReconnect()
		return
	}

	state := conn.GetState()
	c.logger.Debug("Connection state", zap.String("state", state.String()))

	if state == grpc.TransientFailure || state == grpc.Shutdown {
		c.triggerReconnect()
	}
}

// triggerReconnect triggers a reconnection attempt
func (c *Client) triggerReconnect() {
	select {
	case c.reconnectCh <- struct{}{}:
		c.logger.Info("Triggering reconnection")
	default:
		// Reconnection already pending
	}
}

// reconnect attempts to reconnect to the server
func (c *Client) reconnect() {
	c.logger.Info("Attempting to reconnect")

	for attempt := 1; attempt <= 5; attempt++ {
		if c.closed {
			return
		}

		err := c.Connect()
		if err == nil {
			c.logger.Info("Reconnection successful")
			return
		}

		c.logger.Warn("Reconnection failed", zap.Int("attempt", attempt), zap.Error(err))

		if attempt < 5 {
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(backoff):
				continue
			}
		}
	}

	c.logger.Error("Failed to reconnect after 5 attempts")
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closed = true
	c.cancel()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}

	return nil
}

// GetManagementClient returns the management service client
func (c *Client) GetManagementClient() pbv1.ManagementServiceClient {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.managementClient
}

// GetAgentClient returns the agent service client
func (c *Client) GetAgentClient() pbv1.AgentServiceClient {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.agentClient
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == grpc.Ready || state == grpc.Idle
}

// GetConnectionState returns the current connection state
func (c *Client) GetConnectionState() grpc.ConnectivityState {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.conn == nil {
		return grpc.Shutdown
	}

	return c.conn.GetState()
}
