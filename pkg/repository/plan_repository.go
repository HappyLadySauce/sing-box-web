package repository

import (
	"time"

	"gorm.io/gorm"

	"sing-box-web/pkg/models"
)

// PlanRepository interface defines plan data access methods
type PlanRepository interface {
	// Basic CRUD operations
	Create(plan *models.Plan) error
	GetByID(id uint) (*models.Plan, error)
	GetByName(name string) (*models.Plan, error)
	Update(plan *models.Plan) error
	Delete(id uint) error
	
	// List operations
	List(offset, limit int) ([]*models.Plan, int64, error)
	ListActive(offset, limit int) ([]*models.Plan, int64, error)
	ListPublic(offset, limit int) ([]*models.Plan, int64, error)
	ListByStatus(status models.PlanStatus, offset, limit int) ([]*models.Plan, int64, error)
	Search(query string, offset, limit int) ([]*models.Plan, int64, error)
	
	// Business operations
	GetDefaultPlan() (*models.Plan, error)
	GetAvailablePlans() ([]*models.Plan, error)
	GetRecommendedPlans() ([]*models.Plan, error)
	IncrementUserCount(planID uint) error
	DecrementUserCount(planID uint) error
	UpdateUserCount(planID uint, count int) error
	
	// Plan features
	CreateFeature(feature *models.PlanFeature) error
	GetPlanFeatures(planID uint) ([]*models.PlanFeature, error)
	UpdateFeature(feature *models.PlanFeature) error
	DeleteFeature(featureID uint) error
	
	// Plan node access
	CreateNodeAccess(access *models.PlanNodeAccess) error
	GetPlanNodeAccess(planID uint) ([]*models.PlanNodeAccess, error)
	GetNodeAccessPlans(nodeID uint) ([]*models.PlanNodeAccess, error)
	UpdateNodeAccess(access *models.PlanNodeAccess) error
	DeleteNodeAccess(planID, nodeID uint) error
	HasNodeAccess(planID, nodeID uint) (bool, error)
	
	// Statistics
	GetPlanCount() (int64, error)
	GetActivePlanCount() (int64, error)
	GetPlanStatistics(planID uint) (*PlanStatistics, error)
	GetAllPlanStatistics() ([]*PlanStatistics, error)
	
	// Batch operations
	BatchUpdateStatus(planIDs []uint, status models.PlanStatus) error
	BatchEnable(planIDs []uint) error
	BatchDisable(planIDs []uint) error
	BatchDelete(planIDs []uint) error
}

// PlanStatistics represents plan usage statistics
type PlanStatistics struct {
	PlanID          uint    `json:"plan_id"`
	PlanName        string  `json:"plan_name"`
	TotalUsers      int64   `json:"total_users"`
	ActiveUsers     int64   `json:"active_users"`
	UsagePercentage float64 `json:"usage_percentage"`
	TotalRevenue    int64   `json:"total_revenue"`
	AvgTrafficUsage int64   `json:"avg_traffic_usage"`
}

// planRepository implements PlanRepository interface
type planRepository struct {
	db *gorm.DB
}

// NewPlanRepository creates a new plan repository
func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db: db}
}

// Create creates a new plan
func (r *planRepository) Create(plan *models.Plan) error {
	return r.db.Create(plan).Error
}

// GetByID gets plan by ID
func (r *planRepository) GetByID(id uint) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.Preload("Users").First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetByName gets plan by name
func (r *planRepository) GetByName(name string) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.Where("name = ?", name).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// Update updates plan information
func (r *planRepository) Update(plan *models.Plan) error {
	return r.db.Save(plan).Error
}

// Delete soft deletes a plan
func (r *planRepository) Delete(id uint) error {
	return r.db.Delete(&models.Plan{}, id).Error
}

// List gets plans with pagination
func (r *planRepository) List(offset, limit int) ([]*models.Plan, int64, error) {
	var plans []*models.Plan
	var total int64
	
	// Get total count
	if err := r.db.Model(&models.Plan{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get plans with pagination
	err := r.db.Offset(offset).
		Limit(limit).
		Order("sort_order ASC, created_at DESC").
		Find(&plans).Error
	
	return plans, total, err
}

// ListActive gets active plans with pagination
func (r *planRepository) ListActive(offset, limit int) ([]*models.Plan, int64, error) {
	var plans []*models.Plan
	var total int64
	
	query := r.db.Model(&models.Plan{}).
		Where("status = ? AND is_enabled = ?", models.PlanStatusActive, true)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get plans with pagination
	err := query.Offset(offset).
		Limit(limit).
		Order("sort_order ASC, created_at DESC").
		Find(&plans).Error
	
	return plans, total, err
}

// ListPublic gets public plans with pagination
func (r *planRepository) ListPublic(offset, limit int) ([]*models.Plan, int64, error) {
	var plans []*models.Plan
	var total int64
	
	query := r.db.Model(&models.Plan{}).
		Where("status = ? AND is_enabled = ? AND is_public = ?", 
			models.PlanStatusActive, true, true)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get plans with pagination
	err := query.Offset(offset).
		Limit(limit).
		Order("sort_order ASC, created_at DESC").
		Find(&plans).Error
	
	return plans, total, err
}

// ListByStatus gets plans by status with pagination
func (r *planRepository) ListByStatus(status models.PlanStatus, offset, limit int) ([]*models.Plan, int64, error) {
	var plans []*models.Plan
	var total int64
	
	query := r.db.Model(&models.Plan{}).Where("status = ?", status)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get plans with pagination
	err := query.Offset(offset).
		Limit(limit).
		Order("sort_order ASC, created_at DESC").
		Find(&plans).Error
	
	return plans, total, err
}

// Search searches plans by name or description
func (r *planRepository) Search(query string, offset, limit int) ([]*models.Plan, int64, error) {
	var plans []*models.Plan
	var total int64
	
	searchQuery := "%" + query + "%"
	dbQuery := r.db.Model(&models.Plan{}).Where(
		"name LIKE ? OR description LIKE ?",
		searchQuery, searchQuery,
	)
	
	// Get total count
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get plans with pagination
	err := dbQuery.Offset(offset).
		Limit(limit).
		Order("sort_order ASC, created_at DESC").
		Find(&plans).Error
	
	return plans, total, err
}

// GetDefaultPlan gets the default plan (usually the first free plan)
func (r *planRepository) GetDefaultPlan() (*models.Plan, error) {
	var plan models.Plan
	err := r.db.Where("status = ? AND is_enabled = ? AND price = 0", 
		models.PlanStatusActive, true).
		Order("sort_order ASC").
		First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetAvailablePlans gets all available plans for subscription
func (r *planRepository) GetAvailablePlans() ([]*models.Plan, error) {
	var plans []*models.Plan
	now := time.Now()
	
	err := r.db.Where(`
		status = ? AND is_enabled = ? AND is_public = ?
		AND (valid_from IS NULL OR valid_from <= ?)
		AND (valid_until IS NULL OR valid_until >= ?)
		AND (max_users = 0 OR current_users < max_users)
	`, models.PlanStatusActive, true, true, now, now).
		Order("sort_order ASC").
		Find(&plans).Error
	
	return plans, err
}

// GetRecommendedPlans gets recommended plans
func (r *planRepository) GetRecommendedPlans() ([]*models.Plan, error) {
	var plans []*models.Plan
	
	err := r.db.Where("status = ? AND is_enabled = ? AND is_public = ? AND is_recommended = ?", 
		models.PlanStatusActive, true, true, true).
		Order("sort_order ASC").
		Find(&plans).Error
	
	return plans, err
}

// IncrementUserCount increments plan user count
func (r *planRepository) IncrementUserCount(planID uint) error {
	return r.db.Model(&models.Plan{}).
		Where("id = ?", planID).
		UpdateColumn("current_users", gorm.Expr("current_users + 1")).
		Error
}

// DecrementUserCount decrements plan user count
func (r *planRepository) DecrementUserCount(planID uint) error {
	return r.db.Model(&models.Plan{}).
		Where("id = ? AND current_users > 0", planID).
		UpdateColumn("current_users", gorm.Expr("current_users - 1")).
		Error
}

// UpdateUserCount updates plan user count
func (r *planRepository) UpdateUserCount(planID uint, count int) error {
	return r.db.Model(&models.Plan{}).
		Where("id = ?", planID).
		Update("current_users", count).
		Error
}

// CreateFeature creates a new plan feature
func (r *planRepository) CreateFeature(feature *models.PlanFeature) error {
	return r.db.Create(feature).Error
}

// GetPlanFeatures gets features for a plan
func (r *planRepository) GetPlanFeatures(planID uint) ([]*models.PlanFeature, error) {
	var features []*models.PlanFeature
	err := r.db.Where("plan_id = ? AND is_visible = ?", planID, true).
		Order("sort_order ASC").
		Find(&features).Error
	return features, err
}

// UpdateFeature updates plan feature
func (r *planRepository) UpdateFeature(feature *models.PlanFeature) error {
	return r.db.Save(feature).Error
}

// DeleteFeature soft deletes a plan feature
func (r *planRepository) DeleteFeature(featureID uint) error {
	return r.db.Delete(&models.PlanFeature{}, featureID).Error
}

// CreateNodeAccess creates plan node access
func (r *planRepository) CreateNodeAccess(access *models.PlanNodeAccess) error {
	return r.db.Create(access).Error
}

// GetPlanNodeAccess gets node access settings for a plan
func (r *planRepository) GetPlanNodeAccess(planID uint) ([]*models.PlanNodeAccess, error) {
	var access []*models.PlanNodeAccess
	err := r.db.Preload("Node").
		Where("plan_id = ? AND is_enabled = ?", planID, true).
		Order("priority ASC").
		Find(&access).Error
	return access, err
}

// GetNodeAccessPlans gets plans that have access to a node
func (r *planRepository) GetNodeAccessPlans(nodeID uint) ([]*models.PlanNodeAccess, error) {
	var access []*models.PlanNodeAccess
	err := r.db.Preload("Plan").
		Where("node_id = ? AND is_enabled = ?", nodeID, true).
		Order("priority ASC").
		Find(&access).Error
	return access, err
}

// UpdateNodeAccess updates plan node access
func (r *planRepository) UpdateNodeAccess(access *models.PlanNodeAccess) error {
	return r.db.Save(access).Error
}

// DeleteNodeAccess removes plan node access
func (r *planRepository) DeleteNodeAccess(planID, nodeID uint) error {
	return r.db.Where("plan_id = ? AND node_id = ?", planID, nodeID).
		Delete(&models.PlanNodeAccess{}).Error
}

// HasNodeAccess checks if plan has access to a node
func (r *planRepository) HasNodeAccess(planID, nodeID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.PlanNodeAccess{}).
		Where("plan_id = ? AND node_id = ? AND is_enabled = ?", planID, nodeID, true).
		Count(&count).Error
	return count > 0, err
}

// GetPlanCount gets total plan count
func (r *planRepository) GetPlanCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Plan{}).Count(&count).Error
	return count, err
}

// GetActivePlanCount gets active plan count
func (r *planRepository) GetActivePlanCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Plan{}).
		Where("status = ? AND is_enabled = ?", models.PlanStatusActive, true).
		Count(&count).Error
	return count, err
}

// GetPlanStatistics gets statistics for a specific plan
func (r *planRepository) GetPlanStatistics(planID uint) (*PlanStatistics, error) {
	var stats PlanStatistics
	
	// Get plan basic info
	var plan models.Plan
	if err := r.db.First(&plan, planID).Error; err != nil {
		return nil, err
	}
	
	stats.PlanID = plan.ID
	stats.PlanName = plan.Name
	stats.TotalUsers = int64(plan.CurrentUsers)
	
	// Get active users count
	r.db.Model(&models.User{}).
		Where("plan_id = ? AND status = ?", planID, models.UserStatusActive).
		Count(&stats.ActiveUsers)
	
	// Calculate usage percentage
	if plan.MaxUsers > 0 {
		stats.UsagePercentage = float64(plan.CurrentUsers) / float64(plan.MaxUsers) * 100
	}
	
	// Calculate total revenue (simplified - assumes all users pay full price)
	stats.TotalRevenue = int64(plan.CurrentUsers) * plan.Price
	
	// Get average traffic usage
	var avgTraffic struct {
		Avg int64
	}
	r.db.Model(&models.User{}).
		Select("COALESCE(AVG(traffic_used), 0) as avg").
		Where("plan_id = ?", planID).
		Scan(&avgTraffic)
	stats.AvgTrafficUsage = avgTraffic.Avg
	
	return &stats, nil
}

// GetAllPlanStatistics gets statistics for all plans
func (r *planRepository) GetAllPlanStatistics() ([]*PlanStatistics, error) {
	var plans []*models.Plan
	if err := r.db.Find(&plans).Error; err != nil {
		return nil, err
	}
	
	var allStats []*PlanStatistics
	for _, plan := range plans {
		stats, err := r.GetPlanStatistics(plan.ID)
		if err != nil {
			continue // Skip plans with errors
		}
		allStats = append(allStats, stats)
	}
	
	return allStats, nil
}

// BatchUpdateStatus updates status for multiple plans
func (r *planRepository) BatchUpdateStatus(planIDs []uint, status models.PlanStatus) error {
	return r.db.Model(&models.Plan{}).
		Where("id IN ?", planIDs).
		Update("status", status).
		Error
}

// BatchEnable enables multiple plans
func (r *planRepository) BatchEnable(planIDs []uint) error {
	return r.db.Model(&models.Plan{}).
		Where("id IN ?", planIDs).
		Update("is_enabled", true).
		Error
}

// BatchDisable disables multiple plans
func (r *planRepository) BatchDisable(planIDs []uint) error {
	return r.db.Model(&models.Plan{}).
		Where("id IN ?", planIDs).
		Update("is_enabled", false).
		Error
}

// BatchDelete soft deletes multiple plans
func (r *planRepository) BatchDelete(planIDs []uint) error {
	return r.db.Delete(&models.Plan{}, planIDs).Error
}