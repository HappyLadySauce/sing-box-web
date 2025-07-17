package api

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"go.uber.org/zap"

	configv1 "sing-box-web/pkg/config/v1"
	"sing-box-web/pkg/logger"
	pbv1 "sing-box-web/pkg/pb/v1"
)

// Server represents the gRPC API server
type Server struct {
	config     configv1.APIConfig
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *zap.Logger
	
	// Services
	managementService *ManagementService
	agentService     *AgentService
}

// NewServer creates a new gRPC API server
func NewServer(config configv1.APIConfig) (*Server, error) {
	logger := logger.GetLogger().Named("api-server")
	
	// Create gRPC server with options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(config.GRPC.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.GRPC.MaxSendMsgSize),
		grpc.ConnectionTimeout(config.GRPC.ConnectionTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    config.GRPC.KeepaliveTime,
			Timeout: config.GRPC.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.GRPC.KeepaliveTime / 2,
			PermitWithoutStream: true,
		}),
	}
	
	// Add TLS if enabled
	if config.GRPC.TLSEnabled {
		// TODO: Add TLS configuration
	}
	
	grpcServer := grpc.NewServer(opts...)
	
	// Create services
	managementService := NewManagementService(config, logger)
	agentService := NewAgentService(config, logger)
	
	// Register services
	pbv1.RegisterManagementServiceServer(grpcServer, managementService)
	pbv1.RegisterAgentServiceServer(grpcServer, agentService)
	
	// Register reflection service for development
	reflection.Register(grpcServer)
	
	return &Server{
		config:            config,
		grpcServer:        grpcServer,
		logger:            logger,
		managementService: managementService,
		agentService:      agentService,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	// Create listener
	address := fmt.Sprintf("%s:%d", s.config.GRPC.Address, s.config.GRPC.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}
	
	s.listener = listener
	s.logger.Info("gRPC server starting",
		zap.String("address", address),
		zap.Bool("tls", s.config.GRPC.TLSEnabled),
	)
	
	// Start server in goroutine
	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			s.logger.Error("gRPC server failed", zap.Error(err))
		}
	}()
	
	// Start services
	if err := s.managementService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start management service: %w", err)
	}
	
	if err := s.agentService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start agent service: %w", err)
	}
	
	s.logger.Info("gRPC server started successfully")
	return nil
}

// Stop stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("gRPC server stopping")
	
	// Stop services
	if err := s.managementService.Stop(ctx); err != nil {
		s.logger.Error("failed to stop management service", zap.Error(err))
	}
	
	if err := s.agentService.Stop(ctx); err != nil {
		s.logger.Error("failed to stop agent service", zap.Error(err))
	}
	
	// Graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	
	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")
	case <-time.After(30 * time.Second):
		s.logger.Warn("gRPC server force stopped due to timeout")
		s.grpcServer.Stop()
	}
	
	if s.listener != nil {
		s.listener.Close()
	}
	
	return nil
}

// GetAddress returns the server listen address
func (s *Server) GetAddress() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return fmt.Sprintf("%s:%d", s.config.GRPC.Address, s.config.GRPC.Port)
}

// IsHealthy returns true if the server is healthy
func (s *Server) IsHealthy() bool {
	return s.listener != nil && s.grpcServer != nil
}

// GetMetrics returns server metrics
func (s *Server) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"address":     s.GetAddress(),
		"healthy":     s.IsHealthy(),
		"tls_enabled": s.config.GRPC.TLSEnabled,
	}
}