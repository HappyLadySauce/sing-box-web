package api

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"sing-box-web/pkg/database"
	"sing-box-web/pkg/models"
	pbv1 "sing-box-web/pkg/pb/v1"
)

// ManagementService implements the ManagementService gRPC service
type ManagementService struct {
	pbv1.UnimplementedManagementServiceServer

	dbService *database.Service
	logger    *zap.Logger
}

// NewManagementService creates a new ManagementService instance
func NewManagementService(dbService *database.Service, logger *zap.Logger) *ManagementService {
	return &ManagementService{
		dbService: dbService,
		logger:    logger.Named("management-service"),
	}
}

// Start starts the management service
func (s *ManagementService) Start(ctx context.Context) error {
	s.logger.Info("management service starting")
	return nil
}

// Stop stops the management service
func (s *ManagementService) Stop(ctx context.Context) error {
	s.logger.Info("management service stopping")
	return nil
}

// Node management methods

func (s *ManagementService) ListNodes(ctx context.Context, req *pbv1.ListNodesRequest) (*pbv1.ListNodesResponse, error) {
	s.logger.Debug("ListNodes called", zap.Any("request", req))

	// Set default values
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get nodes from database
	nodes, total, err := s.dbService.GetRepository().Node.List(int(offset), int(pageSize))
	if err != nil {
		s.logger.Error("Failed to list nodes", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list nodes")
	}

	// Convert to protobuf format
	pbNodes := make([]*pbv1.NodeInfo, len(nodes))
	for i, node := range nodes {
		pbNodes[i] = s.convertNodeToProto(node)
	}

	return &pbv1.ListNodesResponse{
		Nodes:    pbNodes,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ManagementService) GetNode(ctx context.Context, req *pbv1.GetNodeRequest) (*pbv1.GetNodeResponse, error) {
	s.logger.Debug("GetNode called", zap.String("node_id", req.NodeId))

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	// Parse node ID
	nodeID, err := strconv.ParseUint(req.NodeId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid node_id format")
	}

	// Get node from database
	node, err := s.dbService.GetRepository().Node.GetByID(uint(nodeID))
	if err != nil {
		s.logger.Error("Failed to get node", zap.Error(err), zap.String("node_id", req.NodeId))
		return nil, status.Error(codes.NotFound, "node not found")
	}

	return &pbv1.GetNodeResponse{
		Node: s.convertNodeToProto(node),
	}, nil
}

func (s *ManagementService) RemoveNode(ctx context.Context, req *pbv1.RemoveNodeRequest) (*pbv1.RemoveNodeResponse, error) {
	s.logger.Debug("RemoveNode called", zap.String("node_id", req.NodeId))

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	// Parse node ID
	nodeID, err := strconv.ParseUint(req.NodeId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid node_id format")
	}

	// Check if node exists
	node, err := s.dbService.GetRepository().Node.GetByID(uint(nodeID))
	if err != nil {
		return &pbv1.RemoveNodeResponse{
			Success: false,
			Message: "node not found",
		}, nil
	}

	// Delete the node
	err = s.dbService.GetRepository().Node.Delete(node.ID)
	if err != nil {
		s.logger.Error("Failed to delete node", zap.Error(err), zap.String("node_id", req.NodeId))
		return &pbv1.RemoveNodeResponse{
			Success: false,
			Message: "failed to delete node",
		}, nil
	}

	s.logger.Info("Node removed successfully", zap.String("node_id", req.NodeId), zap.String("name", node.Name))

	return &pbv1.RemoveNodeResponse{
		Success: true,
		Message: "node removed successfully",
	}, nil
}

func (s *ManagementService) UpdateNodeConfig(ctx context.Context, req *pbv1.UpdateNodeConfigRequest) (*pbv1.UpdateNodeConfigResponse, error) {
	s.logger.Debug("UpdateNodeConfig called", zap.String("node_id", req.NodeId))

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	// Parse node ID
	nodeID, err := strconv.ParseUint(req.NodeId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid node_id format")
	}

	// Get node from database
	node, err := s.dbService.GetRepository().Node.GetByID(uint(nodeID))
	if err != nil {
		return &pbv1.UpdateNodeConfigResponse{
			Success: false,
			Message: "node not found",
		}, nil
	}

	// Update node configuration
	if req.ConfigContent != "" {
		node.ConfigContent = req.ConfigContent
	}

	// Update node in database
	err = s.dbService.GetRepository().Node.Update(node)
	if err != nil {
		s.logger.Error("Failed to update node config", zap.Error(err))
		return &pbv1.UpdateNodeConfigResponse{
			Success: false,
			Message: "failed to update node config",
		}, nil
	}

	s.logger.Info("Node config updated successfully", zap.String("node_id", req.NodeId))

	return &pbv1.UpdateNodeConfigResponse{
		Success: true,
		Message: "node config updated successfully",
	}, nil
}

// User management methods

func (s *ManagementService) CreateUser(ctx context.Context, req *pbv1.CreateUserRequest) (*pbv1.CreateUserResponse, error) {
	s.logger.Debug("CreateUser called", zap.String("username", req.Username))

	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	// Check if username already exists
	if _, err := s.dbService.GetRepository().User.GetByUsername(req.Username); err == nil {
		return &pbv1.CreateUserResponse{
			Success: false,
			Message: "username already exists",
			User:    nil,
		}, nil
	}

	// Check if email already exists
	if _, err := s.dbService.GetRepository().User.GetByEmail(req.Email); err == nil {
		return &pbv1.CreateUserResponse{
			Success: false,
			Message: "email already exists",
			User:    nil,
		}, nil
	}

	// Set default plan ID if not provided
	planID := uint(1) // Default plan
	if req.PlanId > 0 {
		planID = uint(req.PlanId)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password, // TODO: Hash password
		DisplayName:  req.Username, // Use username as display name
		Status:       models.UserStatusActive,
		PlanID:       planID,
		TrafficQuota: 10737418240, // Default 10GB
		DeviceLimit:  3,           // Default 3 devices
		SpeedLimit:   0,           // No speed limit
	}

	err := s.dbService.GetRepository().User.Create(user)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return &pbv1.CreateUserResponse{
			Success: false,
			Message: "failed to create user",
			User:    nil,
		}, nil
	}

	s.logger.Info("User created successfully", zap.String("username", user.Username), zap.Uint("id", user.ID))

	return &pbv1.CreateUserResponse{
		Success: true,
		Message: "user created successfully",
		User:    s.convertUserToProto(user),
	}, nil
}

func (s *ManagementService) UpdateUser(ctx context.Context, req *pbv1.UpdateUserRequest) (*pbv1.UpdateUserResponse, error) {
	s.logger.Debug("UpdateUser called", zap.String("user_id", req.UserId))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := strconv.ParseUint(req.UserId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get existing user
	user, err := s.dbService.GetRepository().User.GetByID(uint(userID))
	if err != nil {
		return &pbv1.UpdateUserResponse{
			Success: false,
			Message: "user not found",
			User:    nil,
		}, nil
	}

	// Update user fields
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Username != "" {
		user.Username = req.Username
		user.DisplayName = req.Username // Update display name with username
	}
	if req.PlanId > 0 {
		user.PlanID = uint(req.PlanId)
	}
	if req.Status != "" {
		user.Status = models.UserStatus(req.Status)
	}
	if req.Password != "" {
		user.Password = req.Password // TODO: Hash password
	}

	// Update user in database
	err = s.dbService.GetRepository().User.Update(user)
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return &pbv1.UpdateUserResponse{
			Success: false,
			Message: "failed to update user",
			User:    nil,
		}, nil
	}

	s.logger.Info("User updated successfully", zap.String("user_id", req.UserId), zap.String("username", user.Username))

	return &pbv1.UpdateUserResponse{
		Success: true,
		Message: "user updated successfully",
		User:    s.convertUserToProto(user),
	}, nil
}

func (s *ManagementService) DeleteUser(ctx context.Context, req *pbv1.DeleteUserRequest) (*pbv1.DeleteUserResponse, error) {
	s.logger.Debug("DeleteUser called", zap.String("user_id", req.UserId))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := strconv.ParseUint(req.UserId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Check if user exists
	user, err := s.dbService.GetRepository().User.GetByID(uint(userID))
	if err != nil {
		return &pbv1.DeleteUserResponse{
			Success: false,
			Message: "user not found",
		}, nil
	}

	// Delete user
	err = s.dbService.GetRepository().User.Delete(user.ID)
	if err != nil {
		s.logger.Error("Failed to delete user", zap.Error(err))
		return &pbv1.DeleteUserResponse{
			Success: false,
			Message: "failed to delete user",
		}, nil
	}

	s.logger.Info("User deleted successfully", zap.String("user_id", req.UserId), zap.String("username", user.Username))

	return &pbv1.DeleteUserResponse{
		Success: true,
		Message: "user deleted successfully",
	}, nil
}

func (s *ManagementService) GetUser(ctx context.Context, req *pbv1.GetUserRequest) (*pbv1.GetUserResponse, error) {
	s.logger.Debug("GetUser called", zap.String("user_id", req.UserId))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := strconv.ParseUint(req.UserId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get user from database
	user, err := s.dbService.GetRepository().User.GetByID(uint(userID))
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pbv1.GetUserResponse{
		User: s.convertUserToProto(user),
	}, nil
}

func (s *ManagementService) ListUsers(ctx context.Context, req *pbv1.ListUsersRequest) (*pbv1.ListUsersResponse, error) {
	s.logger.Debug("ListUsers called", zap.Any("request", req))

	// Set default values
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get users from database
	users, total, err := s.dbService.GetRepository().User.List(int(offset), int(pageSize))
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	// Convert to protobuf format
	pbUsers := make([]*pbv1.UserInfo, len(users))
	for i, user := range users {
		pbUsers[i] = s.convertUserToProto(user)
	}

	return &pbv1.ListUsersResponse{
		Users:    pbUsers,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Traffic statistics methods

func (s *ManagementService) GetUserTraffic(ctx context.Context, req *pbv1.GetUserTrafficRequest) (*pbv1.GetUserTrafficResponse, error) {
	s.logger.Debug("GetUserTraffic called", zap.String("user_id", req.UserId))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := strconv.ParseUint(req.UserId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Parse time range
	var startTime, endTime time.Time
	if req.StartTime != nil {
		startTime = req.StartTime.AsTime()
	} else {
		startTime = time.Now().AddDate(0, 0, -7) // Default to last 7 days
	}
	if req.EndTime != nil {
		endTime = req.EndTime.AsTime()
	} else {
		endTime = time.Now()
	}

	// Get traffic records from database
	records, err := s.dbService.GetRepository().Traffic.GetUserTraffic(uint(userID), startTime, endTime)
	if err != nil {
		s.logger.Error("Failed to get user traffic", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user traffic")
	}

	// Calculate totals
	var totalUpload, totalDownload int64
	for _, record := range records {
		totalUpload += record.Upload
		totalDownload += record.Download
	}

	return &pbv1.GetUserTrafficResponse{
		TrafficData:   s.convertTrafficToProto(records),
		TotalUpload:   totalUpload,
		TotalDownload: totalDownload,
	}, nil
}

func (s *ManagementService) GetNodeTraffic(ctx context.Context, req *pbv1.GetNodeTrafficRequest) (*pbv1.GetNodeTrafficResponse, error) {
	s.logger.Debug("GetNodeTraffic called", zap.String("node_id", req.NodeId))

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	// Parse node ID
	nodeID, err := strconv.ParseUint(req.NodeId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid node_id format")
	}

	// Parse time range
	var startTime, endTime time.Time
	if req.StartTime != nil {
		startTime = req.StartTime.AsTime()
	} else {
		startTime = time.Now().AddDate(0, 0, -7) // Default to last 7 days
	}
	if req.EndTime != nil {
		endTime = req.EndTime.AsTime()
	} else {
		endTime = time.Now()
	}

	// Get traffic records from database
	records, err := s.dbService.GetRepository().Traffic.GetNodeTraffic(uint(nodeID), startTime, endTime)
	if err != nil {
		s.logger.Error("Failed to get node traffic", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get node traffic")
	}

	// Calculate totals
	var totalUpload, totalDownload int64
	for _, record := range records {
		totalUpload += record.Upload
		totalDownload += record.Download
	}

	return &pbv1.GetNodeTrafficResponse{
		TrafficData:   s.convertTrafficToProto(records),
		TotalUpload:   totalUpload,
		TotalDownload: totalDownload,
	}, nil
}

// Monitoring data methods

func (s *ManagementService) GetNodeMetrics(ctx context.Context, req *pbv1.GetNodeMetricsRequest) (*pbv1.GetNodeMetricsResponse, error) {
	s.logger.Debug("GetNodeMetrics called", zap.String("node_id", req.NodeId))

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	// Parse node ID
	nodeID, err := strconv.ParseUint(req.NodeId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid node_id format")
	}

	// Get node from database
	node, err := s.dbService.GetRepository().Node.GetByID(uint(nodeID))
	if err != nil {
		s.logger.Error("Failed to get node", zap.Error(err), zap.String("node_id", req.NodeId))
		return nil, status.Error(codes.NotFound, "node not found")
	}

	// Create current metrics from node data
	timestamp := timestamppb.New(time.Now())
	currentMetrics := &pbv1.MetricsData{
		Timestamp:   timestamp,
		CpuUsage:    node.CPUUsage,
		MemoryUsage: node.MemoryUsage,
		DiskUsage:   node.DiskUsage,
		NetworkIn:   node.UploadTraffic,
		NetworkOut:  node.DownloadTraffic,
		Connections: int32(node.CurrentUsers),
	}

	// TODO: Get historical metrics data from time series database
	metricsData := []*pbv1.MetricsData{currentMetrics}

	return &pbv1.GetNodeMetricsResponse{
		MetricsData:    metricsData,
		CurrentMetrics: &pbv1.NodeMetricsInfo{
			CpuUsagePercent:       node.CPUUsage,
			MemoryUsagePercent:    node.MemoryUsage,
			DiskUsagePercent:      node.DiskUsage,
			NetworkInBytesPerSec:  0, // TODO: Calculate per-second rates
			NetworkOutBytesPerSec: 0, // TODO: Calculate per-second rates
			ActiveConnections:     int32(node.CurrentUsers),
			LoadAverage:           node.Load1,
			Timestamp:             timestamp,
		},
	}, nil
}

func (s *ManagementService) GetSystemOverview(ctx context.Context, req *emptypb.Empty) (*pbv1.GetSystemOverviewResponse, error) {
	s.logger.Debug("GetSystemOverview called")

	// Get system statistics
	stats, err := s.dbService.GetRepository().User.GetSystemStats()
	if err != nil {
		s.logger.Error("Failed to get system stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get system stats")
	}

	// Get node statistics
	nodeStats, err := s.dbService.GetRepository().Node.GetNodeStats()
	if err != nil {
		s.logger.Error("Failed to get node stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get node stats")
	}

	// Get today's traffic
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)
	todayTraffic, err := s.dbService.GetRepository().Traffic.GetTotalTrafficInRange(today, tomorrow)
	if err != nil {
		s.logger.Error("Failed to get today's traffic", zap.Error(err))
		todayTraffic = 0
	}

	// Get all nodes for summary
	nodes, _, err := s.dbService.GetRepository().Node.List(0, 100) // Get first 100 nodes
	if err != nil {
		s.logger.Error("Failed to get nodes for summary", zap.Error(err))
		nodes = []*models.Node{}
	}

	// Convert nodes to node summaries
	nodeSummaries := make([]*pbv1.NodeSummary, len(nodes))
	var totalCPU, totalMemory float64
	for i, node := range nodes {
		nodeSummaries[i] = &pbv1.NodeSummary{
			NodeId:          strconv.FormatUint(uint64(node.ID), 10),
			NodeName:        node.Name,
			Status:          string(node.Status),
			UserCount:       int32(node.CurrentUsers),
			ConnectionCount: int32(node.CurrentUsers), // TODO: Get actual connection count
			CpuUsage:        node.CPUUsage,
		}
		totalCPU += node.CPUUsage
		totalMemory += node.MemoryUsage
	}

	// Calculate averages
	var avgCPU, avgMemory float64
	if len(nodes) > 0 {
		avgCPU = totalCPU / float64(len(nodes))
		avgMemory = totalMemory / float64(len(nodes))
	}

	return &pbv1.GetSystemOverviewResponse{
		Stats: &pbv1.SystemStats{
			TotalNodes:        int32(nodeStats.TotalNodes),
			OnlineNodes:       int32(nodeStats.OnlineNodes),
			TotalUsers:        int32(stats.TotalUsers),
			ActiveUsers:       int32(stats.ActiveUsers),
			TotalTrafficToday: todayTraffic,
			TotalConnections:  0, // TODO: Get connection count
			AvgCpuUsage:       avgCPU,
			AvgMemoryUsage:    avgMemory,
		},
		NodeSummaries: nodeSummaries,
		RecentAlerts:  []*pbv1.AlertInfo{}, // TODO: Implement alerts
	}, nil
}

// Configuration management methods

func (s *ManagementService) UpdateGlobalConfig(ctx context.Context, req *pbv1.UpdateGlobalConfigRequest) (*pbv1.UpdateGlobalConfigResponse, error) {
	s.logger.Debug("UpdateGlobalConfig called", zap.String("version", req.Version))

	// TODO: Implement global config update logic with proper storage
	// For now, just return success
	newVersion := req.Version
	if newVersion == "" {
		newVersion = time.Now().Format("20060102150405")
	}

	s.logger.Info("Global config updated", zap.String("version", newVersion))

	return &pbv1.UpdateGlobalConfigResponse{
		Success:    true,
		Message:    "global config updated successfully",
		NewVersion: newVersion,
	}, nil
}

func (s *ManagementService) GetGlobalConfig(ctx context.Context, req *emptypb.Empty) (*pbv1.GetGlobalConfigResponse, error) {
	s.logger.Debug("GetGlobalConfig called")

	// TODO: Implement global config retrieval logic with proper storage
	// For now, return default config
	config := map[string]string{
		"log_level":         "info",
		"max_connections":   "1000",
		"traffic_limit":     "1TB",
		"heartbeat_interval": "30s",
		"backup_enabled":    "true",
	}

	return &pbv1.GetGlobalConfigResponse{
		Config:  config,
		Version: "1.0.0",
	}, nil
}

// Batch operations

func (s *ManagementService) BatchUserOperation(ctx context.Context, req *pbv1.BatchUserOperationRequest) (*pbv1.BatchUserOperationResponse, error) {
	s.logger.Debug("BatchUserOperation called",
		zap.String("operation", req.Operation.String()),
		zap.Int("user_count", len(req.UserIds)),
	)

	if len(req.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_ids is required")
	}

	results := make([]*pbv1.OperationResult, len(req.UserIds))
	successCount := 0

	for i, userID := range req.UserIds {
		// Parse user ID
		id, err := strconv.ParseUint(userID, 10, 32)
		if err != nil {
			results[i] = &pbv1.OperationResult{
				UserId:  userID,
				Success: false,
				Message: "invalid user ID format",
			}
			continue
		}

		// Perform operation based on type
		switch req.Operation {
		case pbv1.BatchUserOperationRequest_DISABLE:
			err = s.dbService.GetRepository().User.UpdateStatus(uint(id), models.UserStatusSuspended)
		case pbv1.BatchUserOperationRequest_ENABLE:
			err = s.dbService.GetRepository().User.UpdateStatus(uint(id), models.UserStatusActive)
		case pbv1.BatchUserOperationRequest_RESET_TRAFFIC:
			err = s.dbService.GetRepository().User.ResetTraffic(uint(id))
		case pbv1.BatchUserOperationRequest_DELETE:
			err = s.dbService.GetRepository().User.Delete(uint(id))
		default:
			err = status.Error(codes.InvalidArgument, "unsupported operation")
		}

		if err != nil {
			results[i] = &pbv1.OperationResult{
				UserId:  userID,
				Success: false,
				Message: err.Error(),
			}
		} else {
			results[i] = &pbv1.OperationResult{
				UserId:  userID,
				Success: true,
				Message: "operation completed successfully",
			}
			successCount++
		}
	}

	s.logger.Info("Batch operation completed",
		zap.String("operation", req.Operation.String()),
		zap.Int("success_count", successCount),
		zap.Int("total_count", len(req.UserIds)),
	)

	return &pbv1.BatchUserOperationResponse{
		Success: successCount > 0,
		Message: fmt.Sprintf("%d/%d operations completed successfully", successCount, len(req.UserIds)),
		Results: results,
	}, nil
}

// Helper functions for converting between models and protobuf

func (s *ManagementService) convertNodeToProto(node *models.Node) *pbv1.NodeInfo {
	var lastSeen *timestamppb.Timestamp
	if node.LastHeartbeat != nil {
		lastSeen = timestamppb.New(*node.LastHeartbeat)
	}

	return &pbv1.NodeInfo{
		NodeId:        strconv.FormatUint(uint64(node.ID), 10),
		NodeName:      node.Name,
		NodeIp:        node.Host,
		Status:        string(node.Status),
		Version:       node.SingBoxVersion,
		LastSeen:      lastSeen,
		UserCount:     int32(node.CurrentUsers),
		ConfigVersion: strconv.Itoa(node.ConfigVersion),
	}
}

func (s *ManagementService) convertUserToProto(user *models.User) *pbv1.UserInfo {
	var expiresAt *timestamppb.Timestamp
	if user.ExpiresAt != nil {
		expiresAt = timestamppb.New(*user.ExpiresAt)
	}

	createdAt := timestamppb.New(user.CreatedAt)
	updatedAt := timestamppb.New(user.UpdatedAt)

	return &pbv1.UserInfo{
		UserId:    strconv.FormatUint(uint64(user.ID), 10),
		Username:  user.Username,
		Email:     user.Email,
		Status:    string(user.Status),
		PlanId:    int64(user.PlanID),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ExpiresAt: expiresAt,
	}
}

func (s *ManagementService) convertTrafficToProto(records []*models.TrafficRecord) []*pbv1.TrafficData {
	pbRecords := make([]*pbv1.TrafficData, len(records))
	for i, record := range records {
		pbRecords[i] = &pbv1.TrafficData{
			Timestamp:     timestamppb.New(record.CreatedAt),
			UploadBytes:   record.Upload,
			DownloadBytes: record.Download,
			NodeId:        strconv.FormatUint(uint64(record.NodeID), 10),
		}
	}
	return pbRecords
}
