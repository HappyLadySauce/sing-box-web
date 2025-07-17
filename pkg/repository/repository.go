package repository

import (
	"time"

	"gorm.io/gorm"

	"sing-box-web/pkg/models"
)

// Manager represents the repository manager
type Manager struct {
	db *gorm.DB
	
	// Repository instances
	User    UserRepository
	Node    NodeRepository
	Plan    PlanRepository
	Traffic TrafficRepository
}

// NewManager creates a new repository manager
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:      db,
		User:    NewUserRepository(db),
		Node:    NewNodeRepository(db),
		Plan:    NewPlanRepository(db),
		Traffic: NewTrafficRepository(db),
	}
}

// GetDB returns the underlying database instance
func (m *Manager) GetDB() *gorm.DB {
	return m.db
}

// Transaction executes a function within a database transaction
func (m *Manager) Transaction(fn func(*gorm.DB) error) error {
	return m.db.Transaction(fn)
}

// Health checks the database connection
func (m *Manager) Health() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close closes the database connection
func (m *Manager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetStatistics returns combined repository statistics
func (m *Manager) GetStatistics() (*models.Statistics, error) {
	stats := &models.Statistics{}
	
	// Get user statistics
	totalUsers, err := m.User.GetUserCount()
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = totalUsers
	
	activeUsers, err := m.User.GetActiveUserCount()
	if err != nil {
		return nil, err
	}
	stats.ActiveUsers = activeUsers
	
	// Get node statistics
	totalNodes, err := m.Node.GetNodeCount()
	if err != nil {
		return nil, err
	}
	stats.TotalNodes = totalNodes
	
	onlineNodes, err := m.Node.GetOnlineNodeCount()
	if err != nil {
		return nil, err
	}
	stats.OnlineNodes = onlineNodes
	
	// Get plan statistics
	totalPlans, err := m.Plan.GetPlanCount()
	if err != nil {
		return nil, err
	}
	stats.TotalPlans = totalPlans
	
	activePlans, err := m.Plan.GetActivePlanCount()
	if err != nil {
		return nil, err
	}
	stats.ActivePlans = activePlans
	
	// Get traffic statistics
	_, _, total, err := m.Traffic.GetTotalTrafficSum(time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	stats.TotalTraffic = total
	
	// Get today's traffic
	today := time.Now().Truncate(24 * time.Hour)
	todayEnd := today.Add(24 * time.Hour)
	_, _, todayTraffic, err := m.Traffic.GetTotalTrafficSum(today, todayEnd)
	if err != nil {
		return nil, err
	}
	stats.TodayTraffic = todayTraffic
	
	// Get monthly traffic
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	_, _, monthlyTraffic, err := m.Traffic.GetTotalTrafficSum(monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	stats.MonthlyTraffic = monthlyTraffic
	
	return stats, nil
}

// InitializeDefaultData creates default data in the database
func (m *Manager) InitializeDefaultData() error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		// Create default plan if not exists
		var planCount int64
		if err := tx.Model(&models.Plan{}).Count(&planCount).Error; err != nil {
			return err
		}
		
		if planCount == 0 {
			defaultPlan := &models.Plan{
				Name:         "Free Plan",
				Description:  "Default free plan with basic features",
				Status:       models.PlanStatusActive,
				Period:       models.PlanPeriodMonthly,
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
				return err
			}
		}
		
		// Create admin user if not exists
		var adminCount int64
		if err := tx.Model(&models.User{}).Where("username = ?", "admin").Count(&adminCount).Error; err != nil {
			return err
		}
		
		if adminCount == 0 {
			// Get the default plan
			var defaultPlan models.Plan
			if err := tx.First(&defaultPlan).Error; err != nil {
				return err
			}
			
			adminUser := &models.User{
				Username:          "admin",
				Email:            "admin@localhost",
				Password:         "$2a$12$example", // This should be properly hashed in production
				DisplayName:      "Administrator",
				Status:           models.UserStatusActive,
				PlanID:           defaultPlan.ID,
				TrafficQuota:     -1, // Unlimited for admin
				DeviceLimit:      10,
				UUID:             generateUUID(),
				SubscriptionToken: generateToken(32),
			}
			if err := tx.Create(adminUser).Error; err != nil {
				return err
			}
		}
		
		return nil
	})
}

// generateUUID generates a UUID (simple implementation for the repository layer)
func generateUUID() string {
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx" // This should use a proper UUID library
}

// generateToken generates a random token (simple implementation)
func generateToken(length int) string {
	return "random-token-placeholder" // This should use proper random generation
}