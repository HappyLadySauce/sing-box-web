package repository

import (
	"time"

	"gorm.io/gorm"

	"sing-box-web/pkg/models"
)

// UserRepository interface defines user data access methods
type UserRepository interface {
	// Basic CRUD operations
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUUID(uuid string) (*models.User, error)
	GetBySubscriptionToken(token string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	
	// List operations
	List(offset, limit int) ([]*models.User, int64, error)
	ListByPlanID(planID uint, offset, limit int) ([]*models.User, int64, error)
	ListByStatus(status models.UserStatus, offset, limit int) ([]*models.User, int64, error)
	Search(query string, offset, limit int) ([]*models.User, int64, error)
	
	// Business operations
	UpdateTrafficUsage(userID uint, upload, download int64) error
	ResetTraffic(userID uint) error
	ResetUserTrafficByPlan(planID uint) error
	UpdateLastLogin(userID uint, ip string) error
	IncrementLoginAttempts(userID uint) error
	ResetLoginAttempts(userID uint) error
	LockUser(userID uint, until time.Time) error
	UnlockUser(userID uint) error
	
	// Statistics
	GetUserCount() (int64, error)
	GetActiveUserCount() (int64, error)
	GetUsersByDateRange(start, end time.Time) ([]*models.User, error)
	GetTopTrafficUsers(limit int) ([]*models.User, error)
	
	// Batch operations
	BatchUpdateStatus(userIDs []uint, status models.UserStatus) error
	BatchDelete(userIDs []uint) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID gets user by ID
func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Plan").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername gets user by username
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Plan").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail gets user by email
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Plan").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUUID gets user by UUID
func (r *userRepository) GetByUUID(uuid string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Plan").Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetBySubscriptionToken gets user by subscription token
func (r *userRepository) GetBySubscriptionToken(token string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Plan").Where("subscription_token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates user information
func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// List gets users with pagination
func (r *userRepository) List(offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	// Get total count
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get users with pagination
	err := r.db.Preload("Plan").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error
	
	return users, total, err
}

// ListByPlanID gets users by plan ID with pagination
func (r *userRepository) ListByPlanID(planID uint, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	query := r.db.Model(&models.User{}).Where("plan_id = ?", planID)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get users with pagination
	err := query.Preload("Plan").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error
	
	return users, total, err
}

// ListByStatus gets users by status with pagination
func (r *userRepository) ListByStatus(status models.UserStatus, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	query := r.db.Model(&models.User{}).Where("status = ?", status)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get users with pagination
	err := query.Preload("Plan").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error
	
	return users, total, err
}

// Search searches users by username, email, or display name
func (r *userRepository) Search(query string, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	searchQuery := "%" + query + "%"
	dbQuery := r.db.Model(&models.User{}).Where(
		"username LIKE ? OR email LIKE ? OR display_name LIKE ?",
		searchQuery, searchQuery, searchQuery,
	)
	
	// Get total count
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get users with pagination
	err := dbQuery.Preload("Plan").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error
	
	return users, total, err
}

// UpdateTrafficUsage updates user traffic usage
func (r *userRepository) UpdateTrafficUsage(userID uint, upload, download int64) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("traffic_used", gorm.Expr("traffic_used + ?", upload+download)).
		Error
}

// ResetTraffic resets user traffic
func (r *userRepository) ResetTraffic(userID uint) error {
	now := time.Now()
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"traffic_used":      0,
			"traffic_reset_date": now.AddDate(0, 1, 0), // Next month
		}).Error
}

// ResetUserTrafficByPlan resets traffic for all users of a specific plan
func (r *userRepository) ResetUserTrafficByPlan(planID uint) error {
	now := time.Now()
	return r.db.Model(&models.User{}).
		Where("plan_id = ?", planID).
		Updates(map[string]interface{}{
			"traffic_used":      0,
			"traffic_reset_date": now.AddDate(0, 1, 0),
		}).Error
}

// UpdateLastLogin updates user last login information
func (r *userRepository) UpdateLastLogin(userID uint, ip string) error {
	now := time.Now()
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
			"login_attempts": 0, // Reset login attempts on successful login
		}).Error
}

// IncrementLoginAttempts increments user login attempts
func (r *userRepository) IncrementLoginAttempts(userID uint) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("login_attempts", gorm.Expr("login_attempts + 1")).
		Error
}

// ResetLoginAttempts resets user login attempts
func (r *userRepository) ResetLoginAttempts(userID uint) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("login_attempts", 0).
		Error
}

// LockUser locks user account until specified time
func (r *userRepository) LockUser(userID uint, until time.Time) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("locked_until", until).
		Error
}

// UnlockUser unlocks user account
func (r *userRepository) UnlockUser(userID uint) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("locked_until", nil).
		Error
}

// GetUserCount gets total user count
func (r *userRepository) GetUserCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// GetActiveUserCount gets active user count
func (r *userRepository) GetActiveUserCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).
		Where("status = ?", models.UserStatusActive).
		Count(&count).Error
	return count, err
}

// GetUsersByDateRange gets users created within date range
func (r *userRepository) GetUsersByDateRange(start, end time.Time) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Preload("Plan").
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// GetTopTrafficUsers gets users with highest traffic usage
func (r *userRepository) GetTopTrafficUsers(limit int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Preload("Plan").
		Order("traffic_used DESC").
		Limit(limit).
		Find(&users).Error
	return users, err
}

// BatchUpdateStatus updates status for multiple users
func (r *userRepository) BatchUpdateStatus(userIDs []uint, status models.UserStatus) error {
	return r.db.Model(&models.User{}).
		Where("id IN ?", userIDs).
		Update("status", status).
		Error
}

// BatchDelete soft deletes multiple users
func (r *userRepository) BatchDelete(userIDs []uint) error {
	return r.db.Delete(&models.User{}, userIDs).Error
}