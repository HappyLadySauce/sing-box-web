package models

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	configv1 "sing-box-web/pkg/config/v1"
)

// Database represents the database connection and operations
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database instance
func NewDatabase(config configv1.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
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

	return &Database{DB: db}, nil
}

// AutoMigrate runs database migrations for all models
func (d *Database) AutoMigrate() error {
	return d.DB.AutoMigrate(
		&User{},
		&Node{},
		&Plan{},
		&PlanFeature{},
		&PlanNodeAccess{},
		&TrafficRecord{},
		&TrafficSummary{},
		&TrafficQuota{},
		&UserNode{},
		&NodeLog{},
	)
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks database connectivity
func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// GetDB returns the underlying GORM DB instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(fn func(*gorm.DB) error) error {
	return d.DB.Transaction(fn)
}

// CreateDefaultData creates default data for the application
func (d *Database) CreateDefaultData() error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		// Create default plan
		var count int64
		tx.Model(&Plan{}).Count(&count)
		if count == 0 {
			defaultPlan := &Plan{
				Name:         "Free Plan",
				Description:  "Default free plan with basic features",
				Status:       PlanStatusActive,
				Period:       PlanPeriodMonthly,
				Price:        0,
				Currency:     "USD",
				TrafficQuota: 10 * 1024 * 1024 * 1024, // 10GB
				SpeedLimit:   0,                        // Unlimited
				DeviceLimit:  1,
				IsPublic:     true,
				IsEnabled:    true,
				Color:        "#2563eb",
				SortOrder:    1,
			}
			if err := tx.Create(defaultPlan).Error; err != nil {
				return fmt.Errorf("failed to create default plan: %w", err)
			}
		}

		// Create admin user if not exists
		var adminCount int64
		tx.Model(&User{}).Where("username = ?", "admin").Count(&adminCount)
		if adminCount == 0 {
			// Get the default plan
			var defaultPlan Plan
			if err := tx.First(&defaultPlan).Error; err != nil {
				return fmt.Errorf("failed to get default plan: %w", err)
			}

			adminUser := &User{
				Username:     "admin",
				Email:        "admin@localhost",
				Password:     "$2a$12$example", // This should be properly hashed
				DisplayName:  "Administrator",
				Status:       UserStatusActive,
				PlanID:       defaultPlan.ID,
				TrafficQuota: -1, // Unlimited for admin
				DeviceLimit:  10,
				UUID:         generateUUID(),
				SubscriptionToken: generateToken(32),
			}
			if err := tx.Create(adminUser).Error; err != nil {
				return fmt.Errorf("failed to create admin user: %w", err)
			}
		}

		return nil
	})
}

// Statistics represents database statistics
type Statistics struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	TotalNodes       int64 `json:"total_nodes"`
	OnlineNodes      int64 `json:"online_nodes"`
	TotalPlans       int64 `json:"total_plans"`
	ActivePlans      int64 `json:"active_plans"`
	TotalTraffic     int64 `json:"total_traffic"`
	TodayTraffic     int64 `json:"today_traffic"`
	MonthlyTraffic   int64 `json:"monthly_traffic"`
}

// GetStatistics returns database statistics
func (d *Database) GetStatistics() (*Statistics, error) {
	stats := &Statistics{}

	// User statistics
	d.DB.Model(&User{}).Count(&stats.TotalUsers)
	d.DB.Model(&User{}).Where("status = ?", UserStatusActive).Count(&stats.ActiveUsers)

	// Node statistics
	d.DB.Model(&Node{}).Count(&stats.TotalNodes)
	d.DB.Model(&Node{}).Where("status = ?", NodeStatusOnline).Count(&stats.OnlineNodes)

	// Plan statistics
	d.DB.Model(&Plan{}).Count(&stats.TotalPlans)
	d.DB.Model(&Plan{}).Where("status = ? AND is_enabled = ?", PlanStatusActive, true).Count(&stats.ActivePlans)

	// Traffic statistics
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var totalTraffic struct {
		Total int64
	}
	d.DB.Model(&TrafficRecord{}).Select("COALESCE(SUM(total), 0) as total").Scan(&totalTraffic)
	stats.TotalTraffic = totalTraffic.Total

	var todayTraffic struct {
		Total int64
	}
	d.DB.Model(&TrafficRecord{}).
		Select("COALESCE(SUM(total), 0) as total").
		Where("record_date = ?", today).
		Scan(&todayTraffic)
	stats.TodayTraffic = todayTraffic.Total

	var monthlyTraffic struct {
		Total int64
	}
	d.DB.Model(&TrafficRecord{}).
		Select("COALESCE(SUM(total), 0) as total").
		Where("record_date >= ?", monthStart).
		Scan(&monthlyTraffic)
	stats.MonthlyTraffic = monthlyTraffic.Total

	return stats, nil
}

// CleanupOldRecords removes old traffic records and logs
func (d *Database) CleanupOldRecords(retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 30 // Default retention
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	return d.DB.Transaction(func(tx *gorm.DB) error {
		// Clean up old traffic records
		if err := tx.Where("created_at < ?", cutoffDate).Delete(&TrafficRecord{}).Error; err != nil {
			return fmt.Errorf("failed to cleanup traffic records: %w", err)
		}

		// Clean up old node logs
		if err := tx.Where("created_at < ?", cutoffDate).Delete(&NodeLog{}).Error; err != nil {
			return fmt.Errorf("failed to cleanup node logs: %w", err)
		}

		return nil
	})
}