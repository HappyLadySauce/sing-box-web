package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	configv1 "sing-box-web/pkg/config/v1"
	"sing-box-web/pkg/logger"
	"sing-box-web/pkg/server/api"
)

// NewAPICommand creates a new API command
func NewAPICommand(ctx context.Context) *cobra.Command {
	var configPath string
	
	cmd := &cobra.Command{
		Use:   "sing-box-api",
		Short: "Sing-box API server",
		Long:  "The sing-box-api provides gRPC API for sing-box-web and sing-box-agent.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, configPath)
		},
	}
	
	cmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")
	
	return cmd
}

func run(ctx context.Context, configPath string) error {
	// Load configuration
	config := configv1.DefaultAPIConfig()
	if configPath != "" {
		// TODO: Load config from file
	}
	
	// Initialize logger
	if err := logger.InitLogger(config.Log); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	log := logger.GetLogger().Named("api-main")
	log.Info("Starting sing-box-api", 
		zap.String("address", config.GRPC.Address),
		zap.Int("port", config.GRPC.Port),
	)
	
	// Create and start API server
	server, err := api.NewServer(*config)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}
	
	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	
	// Wait for context cancellation
	<-ctx.Done()
	
	log.Info("Shutting down sing-box-api")
	return server.Stop(ctx)
}