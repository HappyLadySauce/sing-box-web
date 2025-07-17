package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"go.uber.org/zap"

	configv1 "sing-box-web/pkg/config/v1"
	"sing-box-web/pkg/models"
	"sing-box-web/pkg/repository"
)

// Service represents the database service
type Service struct {
	db         *gorm.DB
	repository *repository.Manager
	logger     *zap.Logger
	config     configv1.DatabaseConfig
}

// New creates a new database service
func New(config configv1.DatabaseConfig, logger *zap.Logger) (*Service, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// Configure GORM
	gormConfig := &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	service := &Service{
		db:         db,
		repository: repository.NewManager(db),
		logger:     logger,
		config:     config,
	}

	return service, nil
}

// AutoMigrate runs database migrations
func (s *Service) AutoMigrate() error {
	s.logger.Info("Starting database migration")
	
	err := s.db.AutoMigrate(
		&models.Plan{},
		&models.PlanFeature{},
		&models.User{},
		&models.Node{},
		&models.UserNode{},
		&models.TrafficRecord{},
		&models.TrafficSummary{},
		&models.TrafficQuota{},
		&models.NodeLog{},
		&models.PlanNodeAccess{},
	)
	
	if err != nil {
		s.logger.Error("Database migration failed", zap.Error(err))
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	s.logger.Info("Database migration completed successfully")
	return nil
}

// InitializeData creates default data
func (s *Service) InitializeData() error {
	s.logger.Info("Initializing default data")
	
	if err := s.repository.InitializeDefaultData(); err != nil {
		s.logger.Error("Failed to initialize default data", zap.Error(err))
		return err
	}
	
	s.logger.Info("Default data initialized successfully")
	return nil
}

// GetRepository returns the repository manager
func (s *Service) GetRepository() *repository.Manager {
	return s.repository
}

// GetDB returns the underlying GORM database instance
func (s *Service) GetDB() *gorm.DB {
	return s.db
}

// Health checks database connectivity
func (s *Service) Health() error {
	return s.repository.Health()
}

// Close closes the database connection
func (s *Service) Close() error {
	s.logger.Info("Closing database connection")
	return s.repository.Close()
}

// GetStatistics returns database statistics
func (s *Service) GetStatistics() (*models.Statistics, error) {
	return s.repository.GetStatistics()
}

// Transaction executes a function within a database transaction
func (s *Service) Transaction(fn func(*gorm.DB) error) error {
	return s.repository.Transaction(fn)
}

// StartMaintenanceTasks starts periodic maintenance tasks
func (s *Service) StartMaintenanceTasks() {
	// Start a goroutine for periodic cleanup
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run daily
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				s.runMaintenanceTasks()
			}
		}
	}()
}

// runMaintenanceTasks executes periodic maintenance tasks
func (s *Service) runMaintenanceTasks() {
	s.logger.Info("Starting maintenance tasks")
	
	// Cleanup old traffic records (keep 30 days)
	if err := s.repository.Traffic.CleanupOldRecords(30); err != nil {
		s.logger.Error("Failed to cleanup old traffic records", zap.Error(err))
	}
	
	// Cleanup old summaries (keep 90 days)
	if err := s.repository.Traffic.CleanupOldSummaries(90); err != nil {
		s.logger.Error("Failed to cleanup old traffic summaries", zap.Error(err))
	}
	
	// Aggregate daily data for yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	if err := s.repository.Traffic.AggregateDailyData(yesterday); err != nil {
		s.logger.Error("Failed to aggregate daily traffic data", zap.Error(err))
	}
	
	// Aggregate monthly data for last month (on the 1st of each month)
	if time.Now().Day() == 1 {
		lastMonth := time.Now().AddDate(0, -1, 0)
		if err := s.repository.Traffic.AggregateMonthlyData(lastMonth); err != nil {
			s.logger.Error("Failed to aggregate monthly traffic data", zap.Error(err))
		}
	}
	
	s.logger.Info("Maintenance tasks completed")
}

