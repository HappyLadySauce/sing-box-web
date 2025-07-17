package agent

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	configv1 "sing-box-web/pkg/config/v1"
	"sing-box-web/pkg/logger"
	pbv1 "sing-box-web/pkg/pb/v1"
)

// Agent represents the sing-box agent
type Agent struct {
	config configv1.AgentConfig
	logger *zap.Logger

	// gRPC client connection
	apiClient pbv1.AgentServiceClient
	conn      *grpc.ClientConn

	// Node management
	nodeInfo     *pbv1.RegisterNodeRequest
	lastSeen     time.Time
	registered   bool
	registeredMu sync.RWMutex

	// Metrics collection
	metricsCollector *MetricsCollector

	// Sing-box management
	singboxManager *SingboxManager

	// Shutdown
	shutdownCtx context.Context
	shutdown    context.CancelFunc
}

// NewAgent creates a new agent instance
func NewAgent(config configv1.AgentConfig) (*Agent, error) {
	logger := logger.GetLogger().Named("agent")

	// Create shutdown context
	shutdownCtx, shutdown := context.WithCancel(context.Background())

	// Create agent
	agent := &Agent{
		config:      config,
		logger:      logger,
		shutdownCtx: shutdownCtx,
		shutdown:    shutdown,
	}

	// Initialize node info
	if err := agent.initializeNodeInfo(); err != nil {
		return nil, fmt.Errorf("failed to initialize node info: %w", err)
	}

	// Create metrics collector
	agent.metricsCollector = NewMetricsCollector(logger)

	// Create sing-box manager
	agent.singboxManager = NewSingboxManager(config, logger)

	return agent, nil
}

// Start starts the agent
func (a *Agent) Start(ctx context.Context) error {
	a.logger.Info("agent starting")

	// Connect to API server
	if err := a.connectToAPI(); err != nil {
		return fmt.Errorf("failed to connect to API server: %w", err)
	}

	// Register node
	if err := a.registerNode(); err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	// Start metrics collection
	if err := a.metricsCollector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start metrics collector: %w", err)
	}

	// Start sing-box manager
	if err := a.singboxManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start sing-box manager: %w", err)
	}

	// Start background tasks
	go a.heartbeatLoop()
	go a.metricsReportLoop()
	go a.trafficReportLoop()
	go a.commandProcessorLoop()

	a.logger.Info("agent started successfully")
	return nil
}

// Stop stops the agent
func (a *Agent) Stop(ctx context.Context) error {
	a.logger.Info("agent stopping")

	// Cancel background tasks
	a.shutdown()

	// Stop sing-box manager
	if err := a.singboxManager.Stop(ctx); err != nil {
		a.logger.Error("failed to stop sing-box manager", zap.Error(err))
	}

	// Stop metrics collector
	if err := a.metricsCollector.Stop(ctx); err != nil {
		a.logger.Error("failed to stop metrics collector", zap.Error(err))
	}

	// Close gRPC connection
	if a.conn != nil {
		a.conn.Close()
	}

	a.logger.Info("agent stopped")
	return nil
}

// initializeNodeInfo initializes the node information
func (a *Agent) initializeNodeInfo() error {
	// Get node IP
	nodeIP, err := a.getNodeIP()
	if err != nil {
		return fmt.Errorf("failed to get node IP: %w", err)
	}

	// Get node capabilities
	capabilities := a.getNodeCapabilities()

	a.nodeInfo = &pbv1.RegisterNodeRequest{
		NodeId:     a.config.Node.NodeID,
		NodeName:   a.config.Node.NodeName,
		NodeIp:     nodeIP,
		Version:    "1.0.0", // TODO: Get actual version
		Capability: capabilities,
	}

	return nil
}

// getNodeIP gets the node's IP address
func (a *Agent) getNodeIP() (string, error) {
	// Get local IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// getNodeCapabilities gets the node's capabilities
func (a *Agent) getNodeCapabilities() *pbv1.NodeCapability {
	return &pbv1.NodeCapability{
		MaxConnections:    int32(a.config.Node.MaxUsers),
		MaxBandwidthMbps:  1000, // 1Gbps default
		SupportedProtocols: []string{"vless", "vmess", "trojan", "shadowsocks"},
		Features: map[string]string{
			"metrics":        "enabled",
			"traffic_stats":  "enabled",
			"user_management": "enabled",
		},
	}
}

// connectToAPI connects to the API server
func (a *Agent) connectToAPI() error {
	apiAddress := fmt.Sprintf("%s:%d", a.config.APIServer.Address, a.config.APIServer.Port)
	a.logger.Info("connecting to API server", zap.String("address", apiAddress))

	// Create connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, apiAddress, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to API server: %w", err)
	}

	a.conn = conn
	a.apiClient = pbv1.NewAgentServiceClient(conn)

	a.logger.Info("connected to API server successfully")
	return nil
}

// registerNode registers the node with the API server
func (a *Agent) registerNode() error {
	a.logger.Info("registering node", zap.String("node_id", a.nodeInfo.NodeId))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := a.apiClient.RegisterNode(ctx, a.nodeInfo)
	if err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("node registration failed: %s", resp.Message)
	}

	a.registeredMu.Lock()
	a.registered = true
	a.lastSeen = time.Now()
	a.registeredMu.Unlock()

	a.logger.Info("node registered successfully", zap.String("message", resp.Message))
	return nil
}

// heartbeatLoop sends periodic heartbeats to the API server
func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(a.config.Monitor.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.shutdownCtx.Done():
			return
		case <-ticker.C:
			a.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat to the API server
func (a *Agent) sendHeartbeat() {
	a.registeredMu.RLock()
	if !a.registered {
		a.registeredMu.RUnlock()
		return
	}
	a.registeredMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current status
	status := &pbv1.NodeStatus{
		Status:            "online",
		SingBoxVersion:    "1.0.0",
		ActiveConnections: int32(10), // TODO: Get actual connection count
		ErrorMessage:      "",
	}

	req := &pbv1.HeartbeatRequest{
		NodeId: a.nodeInfo.NodeId,
		Status: status,
	}

	resp, err := a.apiClient.Heartbeat(ctx, req)
	if err != nil {
		a.logger.Error("failed to send heartbeat", zap.Error(err))
		return
	}

	if !resp.Success {
		a.logger.Error("heartbeat failed")
		return
	}

	// Process pending commands
	a.processPendingCommands(resp.PendingCommands)

	a.registeredMu.Lock()
	a.lastSeen = time.Now()
	a.registeredMu.Unlock()
}

// metricsReportLoop reports metrics to the API server
func (a *Agent) metricsReportLoop() {
	ticker := time.NewTicker(a.config.Monitor.SystemMetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.shutdownCtx.Done():
			return
		case <-ticker.C:
			a.reportMetrics()
		}
	}
}

// reportMetrics reports current metrics to the API server
func (a *Agent) reportMetrics() {
	a.registeredMu.RLock()
	if !a.registered {
		a.registeredMu.RUnlock()
		return
	}
	a.registeredMu.RUnlock()

	metrics := a.metricsCollector.GetMetrics()
	if metrics == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pbv1.ReportMetricsRequest{
		NodeId:  a.nodeInfo.NodeId,
		Metrics: metrics,
	}

	resp, err := a.apiClient.ReportMetrics(ctx, req)
	if err != nil {
		a.logger.Error("failed to report metrics", zap.Error(err))
		return
	}

	if !resp.Success {
		a.logger.Error("metrics report failed", zap.String("message", resp.Message))
	}
}

// trafficReportLoop reports traffic statistics to the API server
func (a *Agent) trafficReportLoop() {
	ticker := time.NewTicker(a.config.Monitor.TrafficReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.shutdownCtx.Done():
			return
		case <-ticker.C:
			a.reportTraffic()
		}
	}
}

// reportTraffic reports traffic statistics to the API server
func (a *Agent) reportTraffic() {
	a.registeredMu.RLock()
	if !a.registered {
		a.registeredMu.RUnlock()
		return
	}
	a.registeredMu.RUnlock()

	// Get traffic data from sing-box manager
	trafficData := a.singboxManager.GetTrafficData()
	if len(trafficData) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pbv1.ReportTrafficRequest{
		NodeId:      a.nodeInfo.NodeId,
		UserTraffic: trafficData,
	}

	resp, err := a.apiClient.ReportTraffic(ctx, req)
	if err != nil {
		a.logger.Error("failed to report traffic", zap.Error(err))
		return
	}

	if !resp.Success {
		a.logger.Error("traffic report failed", zap.String("message", resp.Message))
	}
}

// commandProcessorLoop processes commands from the API server
func (a *Agent) commandProcessorLoop() {
	// This loop would handle long-running command processing
	// For now, commands are processed in the heartbeat response
	<-a.shutdownCtx.Done()
}

// processPendingCommands processes pending commands from the API server
func (a *Agent) processPendingCommands(commands []*pbv1.PendingCommand) {
	for _, cmd := range commands {
		a.logger.Info("processing command",
			zap.String("command_id", cmd.CommandId),
			zap.String("command_type", cmd.Command.Type.String()),
		)

		switch cmd.Command.Type {
		case pbv1.UserCommand_ADD_USER:
			a.handleAddUser(cmd)
		case pbv1.UserCommand_REMOVE_USER:
			a.handleRemoveUser(cmd)
		case pbv1.UserCommand_UPDATE_USER:
			a.handleUpdateUser(cmd)
		case pbv1.UserCommand_RESET_TRAFFIC:
			a.handleResetTraffic(cmd)
		default:
			a.logger.Warn("unknown command type", zap.String("type", cmd.Command.Type.String()))
		}
	}
}

// handleAddUser handles add user command
func (a *Agent) handleAddUser(cmd *pbv1.PendingCommand) {
	userID := cmd.Command.UserId
	a.logger.Info("adding user", zap.String("user_id", userID))

	// Add user to sing-box configuration
	if err := a.singboxManager.AddUser(userID, cmd.Command.Parameters); err != nil {
		a.logger.Error("failed to add user", zap.Error(err))
		return
	}

	a.logger.Info("user added successfully", zap.String("user_id", userID))
}

// handleRemoveUser handles remove user command
func (a *Agent) handleRemoveUser(cmd *pbv1.PendingCommand) {
	userID := cmd.Command.UserId
	a.logger.Info("removing user", zap.String("user_id", userID))

	// Remove user from sing-box configuration
	if err := a.singboxManager.RemoveUser(userID); err != nil {
		a.logger.Error("failed to remove user", zap.Error(err))
		return
	}

	a.logger.Info("user removed successfully", zap.String("user_id", userID))
}

// handleUpdateUser handles update user command
func (a *Agent) handleUpdateUser(cmd *pbv1.PendingCommand) {
	userID := cmd.Command.UserId
	a.logger.Info("updating user", zap.String("user_id", userID))

	// Update user in sing-box configuration
	if err := a.singboxManager.UpdateUser(userID, cmd.Command.Parameters); err != nil {
		a.logger.Error("failed to update user", zap.Error(err))
		return
	}

	a.logger.Info("user updated successfully", zap.String("user_id", userID))
}

// handleResetTraffic handles reset traffic command
func (a *Agent) handleResetTraffic(cmd *pbv1.PendingCommand) {
	userID := cmd.Command.UserId
	a.logger.Info("resetting traffic", zap.String("user_id", userID))

	// Reset traffic for user
	if err := a.singboxManager.ResetTraffic(userID); err != nil {
		a.logger.Error("failed to reset traffic", zap.Error(err))
		return
	}

	a.logger.Info("traffic reset successfully", zap.String("user_id", userID))
}

// IsRegistered returns true if the node is registered
func (a *Agent) IsRegistered() bool {
	a.registeredMu.RLock()
	defer a.registeredMu.RUnlock()
	return a.registered
}

// GetLastSeen returns the last seen timestamp
func (a *Agent) GetLastSeen() time.Time {
	a.registeredMu.RLock()
	defer a.registeredMu.RUnlock()
	return a.lastSeen
}

// GetNodeInfo returns the node information
func (a *Agent) GetNodeInfo() *pbv1.RegisterNodeRequest {
	return a.nodeInfo
}