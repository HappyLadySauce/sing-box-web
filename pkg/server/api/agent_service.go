package api

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	configv1 "sing-box-web/pkg/config/v1"
	pbv1 "sing-box-web/pkg/pb/v1"
)

// AgentService implements the AgentService gRPC service
type AgentService struct {
	pbv1.UnimplementedAgentServiceServer
	
	config configv1.APIConfig
	logger *zap.Logger
	
	// Node management
	nodes    map[string]*NodeState
	nodesMux sync.RWMutex
	
	// Command queue for nodes
	commandQueues map[string]chan *pbv1.PendingCommand
	queuesMux     sync.RWMutex
}

// NodeState represents the state of a connected node
type NodeState struct {
	Info      *pbv1.RegisterNodeRequest
	LastSeen  time.Time
	Status    *pbv1.NodeStatus
	Metrics   *pbv1.NodeMetrics
}

// NewAgentService creates a new AgentService instance
func NewAgentService(config configv1.APIConfig, logger *zap.Logger) *AgentService {
	return &AgentService{
		config:        config,
		logger:        logger.Named("agent-service"),
		nodes:         make(map[string]*NodeState),
		commandQueues: make(map[string]chan *pbv1.PendingCommand),
	}
}

// Start starts the agent service
func (s *AgentService) Start(ctx context.Context) error {
	s.logger.Info("agent service starting")
	
	// Start cleanup goroutine for offline nodes
	go s.cleanupOfflineNodes(ctx)
	
	return nil
}

// Stop stops the agent service
func (s *AgentService) Stop(ctx context.Context) error {
	s.logger.Info("agent service stopping")
	
	// Close all command queues
	s.queuesMux.Lock()
	for nodeID, queue := range s.commandQueues {
		close(queue)
		delete(s.commandQueues, nodeID)
	}
	s.queuesMux.Unlock()
	
	return nil
}

// RegisterNode registers a new node
func (s *AgentService) RegisterNode(ctx context.Context, req *pbv1.RegisterNodeRequest) (*pbv1.RegisterNodeResponse, error) {
	s.logger.Info("RegisterNode called", 
		zap.String("node_id", req.NodeId),
		zap.String("node_name", req.NodeName),
	)
	
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}
	
	// Update node state
	s.nodesMux.Lock()
	s.nodes[req.NodeId] = &NodeState{
		Info:     req,
		LastSeen: time.Now(),
		Status:   &pbv1.NodeStatus{Status: "online"},
	}
	s.nodesMux.Unlock()
	
	// Create command queue for the node if it doesn't exist
	s.queuesMux.Lock()
	if _, exists := s.commandQueues[req.NodeId]; !exists {
		s.commandQueues[req.NodeId] = make(chan *pbv1.PendingCommand, 100)
	}
	s.queuesMux.Unlock()
	
	s.logger.Info("node registered successfully", zap.String("node_id", req.NodeId))
	
	return &pbv1.RegisterNodeResponse{
		Success: true,
		Message: "node registered successfully",
	}, nil
}

// Heartbeat handles node heartbeat
func (s *AgentService) Heartbeat(ctx context.Context, req *pbv1.HeartbeatRequest) (*pbv1.HeartbeatResponse, error) {
	s.logger.Debug("Heartbeat called", zap.String("node_id", req.NodeId))
	
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}
	
	// Update node last seen time and status
	s.nodesMux.Lock()
	if node, exists := s.nodes[req.NodeId]; exists {
		node.LastSeen = time.Now()
		if req.Status != nil {
			node.Status = req.Status
		}
	} else {
		s.nodesMux.Unlock()
		return nil, status.Error(codes.NotFound, "node not registered")
	}
	s.nodesMux.Unlock()
	
	// Get pending commands
	commands := s.getPendingCommands(req.NodeId)
	
	return &pbv1.HeartbeatResponse{
		Success:         true,
		PendingCommands: commands,
	}, nil
}

// ReportMetrics handles metrics reporting from nodes
func (s *AgentService) ReportMetrics(ctx context.Context, req *pbv1.ReportMetricsRequest) (*pbv1.ReportMetricsResponse, error) {
	s.logger.Debug("ReportMetrics called", zap.String("node_id", req.NodeId))
	
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}
	
	// Update node metrics
	s.nodesMux.Lock()
	if node, exists := s.nodes[req.NodeId]; exists {
		node.Metrics = req.Metrics
		node.LastSeen = time.Now()
	}
	s.nodesMux.Unlock()
	
	// TODO: Store metrics in database
	
	return &pbv1.ReportMetricsResponse{
		Success: true,
		Message: "metrics received",
	}, nil
}

// ReportTraffic handles traffic reporting from nodes
func (s *AgentService) ReportTraffic(ctx context.Context, req *pbv1.ReportTrafficRequest) (*pbv1.ReportTrafficResponse, error) {
	s.logger.Debug("ReportTraffic called", 
		zap.String("node_id", req.NodeId),
		zap.Int("traffic_entries", len(req.UserTraffic)),
	)
	
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}
	
	// TODO: Store traffic data in database
	// TODO: Check traffic limits and generate alerts
	
	return &pbv1.ReportTrafficResponse{
		Success: true,
		Message: "traffic data received",
	}, nil
}

// UpdateConfig handles configuration updates for nodes
func (s *AgentService) UpdateConfig(ctx context.Context, req *pbv1.UpdateConfigRequest) (*pbv1.UpdateConfigResponse, error) {
	s.logger.Debug("UpdateConfig called", zap.String("node_id", req.NodeId))
	
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}
	
	// TODO: Validate and store configuration
	// TODO: Notify node about config update
	
	return &pbv1.UpdateConfigResponse{
		Success:        true,
		Message:        "configuration updated",
		AppliedVersion: req.ConfigVersion,
	}, nil
}

// ExecuteUserCommand handles user management commands
func (s *AgentService) ExecuteUserCommand(ctx context.Context, req *pbv1.ExecuteUserCommandRequest) (*pbv1.ExecuteUserCommandResponse, error) {
	s.logger.Debug("ExecuteUserCommand called", 
		zap.String("node_id", req.NodeId),
		zap.String("command_type", req.Command.Type.String()),
		zap.String("user_id", req.Command.UserId),
	)
	
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}
	
	if req.Command == nil {
		return nil, status.Error(codes.InvalidArgument, "command is required")
	}
	
	// TODO: Execute user command on the specified node
	
	return &pbv1.ExecuteUserCommandResponse{
		Success: false,
		Message: "not implemented",
		Result:  "",
	}, nil
}

// RestartSingBox handles sing-box restart requests
func (s *AgentService) RestartSingBox(ctx context.Context, req *pbv1.RestartSingBoxRequest) (*pbv1.RestartSingBoxResponse, error) {
	s.logger.Debug("RestartSingBox called", zap.String("node_id", req.NodeId))
	
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}
	
	// Add restart command to node's command queue
	command := &pbv1.PendingCommand{
		CommandId: generateCommandID(),
		Command: &pbv1.UserCommand{
			Type:   pbv1.UserCommand_RESET_TRAFFIC, // Use any type for internal commands
			UserId: "system",
			Parameters: map[string]string{
				"action": "restart_singbox",
				"reason": req.Reason,
			},
		},
		CreatedAt: timestamppb.Now(),
	}
	
	if err := s.sendCommandToNode(req.NodeId, command); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send restart command: %v", err)
	}
	
	return &pbv1.RestartSingBoxResponse{
		Success: true,
		Message: "restart command sent",
	}, nil
}

// GetNodeStatus gets the current status of a node
func (s *AgentService) GetNodeStatus(ctx context.Context, req *pbv1.GetNodeStatusRequest) (*pbv1.GetNodeStatusResponse, error) {
	s.logger.Debug("GetNodeStatus called", zap.String("node_id", req.NodeId))
	
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}
	
	s.nodesMux.RLock()
	node, exists := s.nodes[req.NodeId]
	s.nodesMux.RUnlock()
	
	if !exists {
		return nil, status.Error(codes.NotFound, "node not found")
	}
	
	return &pbv1.GetNodeStatusResponse{
		Status:        node.Status,
		Metrics:       node.Metrics,
		ConfigVersion: "", // TODO: Return actual config version
	}, nil
}

// Helper methods

// getPendingCommands gets pending commands for a node
func (s *AgentService) getPendingCommands(nodeID string) []*pbv1.PendingCommand {
	s.queuesMux.RLock()
	queue, exists := s.commandQueues[nodeID]
	s.queuesMux.RUnlock()
	
	if !exists {
		return nil
	}
	
	var commands []*pbv1.PendingCommand
	
	// Non-blocking read of available commands
	for {
		select {
		case cmd := <-queue:
			commands = append(commands, cmd)
		default:
			// No more commands available
			return commands
		}
	}
}

// sendCommandToNode sends a command to a specific node
func (s *AgentService) sendCommandToNode(nodeID string, command *pbv1.PendingCommand) error {
	s.queuesMux.RLock()
	queue, exists := s.commandQueues[nodeID]
	s.queuesMux.RUnlock()
	
	if !exists {
		return status.Errorf(codes.NotFound, "node %s not found", nodeID)
	}
	
	select {
	case queue <- command:
		return nil
	default:
		return status.Errorf(codes.ResourceExhausted, "command queue full for node %s", nodeID)
	}
}

// cleanupOfflineNodes periodically removes offline nodes
func (s *AgentService) cleanupOfflineNodes(ctx context.Context) {
	ticker := time.NewTicker(s.config.Business.Node.ConfigSyncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.performCleanup()
		}
	}
}

// performCleanup removes nodes that haven't been seen for too long
func (s *AgentService) performCleanup() {
	maxOfflineTime := s.config.Business.Node.MaxOfflineTime
	cutoff := time.Now().Add(-maxOfflineTime)
	
	s.nodesMux.Lock()
	for nodeID, node := range s.nodes {
		if node.LastSeen.Before(cutoff) {
			s.logger.Info("removing offline node", 
				zap.String("node_id", nodeID),
				zap.Time("last_seen", node.LastSeen),
			)
			delete(s.nodes, nodeID)
			
			// Close command queue
			s.queuesMux.Lock()
			if queue, exists := s.commandQueues[nodeID]; exists {
				close(queue)
				delete(s.commandQueues, nodeID)
			}
			s.queuesMux.Unlock()
		}
	}
	s.nodesMux.Unlock()
}

// GetNodeStates returns current states of all nodes (for monitoring)
func (s *AgentService) GetNodeStates() map[string]*NodeState {
	s.nodesMux.RLock()
	defer s.nodesMux.RUnlock()
	
	states := make(map[string]*NodeState)
	for nodeID, state := range s.nodes {
		// Create a copy to avoid race conditions
		states[nodeID] = &NodeState{
			Info:     state.Info,
			LastSeen: state.LastSeen,
			Status:   state.Status,
			Metrics:  state.Metrics,
		}
	}
	
	return states
}

// generateCommandID generates a unique command ID
func generateCommandID() string {
	return "cmd-" + time.Now().Format("20060102-150405") + "-" + time.Now().Format("000000")
}