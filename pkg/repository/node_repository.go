package repository

import (
	"time"

	"gorm.io/gorm"

	"sing-box-web/pkg/models"
)

// NodeRepository interface defines node data access methods
type NodeRepository interface {
	// Basic CRUD operations
	Create(node *models.Node) error
	GetByID(id uint) (*models.Node, error)
	GetByName(name string) (*models.Node, error)
	Update(node *models.Node) error
	Delete(id uint) error
	
	// List operations
	List(offset, limit int) ([]*models.Node, int64, error)
	ListByStatus(status models.NodeStatus, offset, limit int) ([]*models.Node, int64, error)
	ListByType(nodeType models.NodeType, offset, limit int) ([]*models.Node, int64, error)
	ListByRegion(region string, offset, limit int) ([]*models.Node, int64, error)
	ListEnabled(offset, limit int) ([]*models.Node, int64, error)
	ListAvailable(offset, limit int) ([]*models.Node, int64, error)
	Search(query string, offset, limit int) ([]*models.Node, int64, error)
	
	// Business operations
	UpdateHeartbeat(nodeID uint) error
	UpdateStatus(nodeID uint, status models.NodeStatus) error
	UpdateSystemInfo(nodeID uint, cpu, memory, disk, load1, load5, load15 float64) error
	UpdateTraffic(nodeID uint, upload, download int64) error
	UpdateUserCount(nodeID uint, count int) error
	IncrementUserCount(nodeID uint) error
	DecrementUserCount(nodeID uint) error
	
	// Statistics
	GetNodeCount() (int64, error)
	GetOnlineNodeCount() (int64, error)
	GetNodesByRegion() (map[string]int64, error)
	GetNodesByType() (map[string]int64, error)
	GetTopTrafficNodes(limit int) ([]*models.Node, error)
	GetNodesWithHighLoad(cpuThreshold, memoryThreshold float64) ([]*models.Node, error)
	GetOfflineNodes(threshold time.Duration) ([]*models.Node, error)
	GetNodeStats() (*models.NodeStats, error)
	
	// Node access management
	GetUserNodes(userID uint) ([]*models.Node, error)
	GetNodeUsers(nodeID uint) ([]*models.User, error)
	AddUserToNode(userID, nodeID uint) error
	RemoveUserFromNode(userID, nodeID uint) error
	SetUserNodePriority(userID, nodeID uint, priority int) error
	EnableUserNode(userID, nodeID uint) error
	DisableUserNode(userID, nodeID uint) error
	
	// Batch operations
	BatchUpdateStatus(nodeIDs []uint, status models.NodeStatus) error
	BatchEnable(nodeIDs []uint) error
	BatchDisable(nodeIDs []uint) error
	BatchDelete(nodeIDs []uint) error
}

// nodeRepository implements NodeRepository interface
type nodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository creates a new node repository
func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepository{db: db}
}

// Create creates a new node
func (r *nodeRepository) Create(node *models.Node) error {
	return r.db.Create(node).Error
}

// GetByID gets node by ID
func (r *nodeRepository) GetByID(id uint) (*models.Node, error) {
	var node models.Node
	err := r.db.First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetByName gets node by name
func (r *nodeRepository) GetByName(name string) (*models.Node, error) {
	var node models.Node
	err := r.db.Where("name = ?", name).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// Update updates node information
func (r *nodeRepository) Update(node *models.Node) error {
	return r.db.Save(node).Error
}

// Delete soft deletes a node
func (r *nodeRepository) Delete(id uint) error {
	return r.db.Delete(&models.Node{}, id).Error
}

// List gets nodes with pagination
func (r *nodeRepository) List(offset, limit int) ([]*models.Node, int64, error) {
	var nodes []*models.Node
	var total int64
	
	// Get total count
	if err := r.db.Model(&models.Node{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get nodes with pagination
	err := r.db.Offset(offset).
		Limit(limit).
		Order("sort ASC, created_at DESC").
		Find(&nodes).Error
	
	return nodes, total, err
}

// ListByStatus gets nodes by status with pagination
func (r *nodeRepository) ListByStatus(status models.NodeStatus, offset, limit int) ([]*models.Node, int64, error) {
	var nodes []*models.Node
	var total int64
	
	query := r.db.Model(&models.Node{}).Where("status = ?", status)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get nodes with pagination
	err := query.Offset(offset).
		Limit(limit).
		Order("sort ASC, created_at DESC").
		Find(&nodes).Error
	
	return nodes, total, err
}

// ListByType gets nodes by type with pagination
func (r *nodeRepository) ListByType(nodeType models.NodeType, offset, limit int) ([]*models.Node, int64, error) {
	var nodes []*models.Node
	var total int64
	
	query := r.db.Model(&models.Node{}).Where("type = ?", nodeType)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get nodes with pagination
	err := query.Offset(offset).
		Limit(limit).
		Order("sort ASC, created_at DESC").
		Find(&nodes).Error
	
	return nodes, total, err
}

// ListByRegion gets nodes by region with pagination
func (r *nodeRepository) ListByRegion(region string, offset, limit int) ([]*models.Node, int64, error) {
	var nodes []*models.Node
	var total int64
	
	query := r.db.Model(&models.Node{}).Where("region = ?", region)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get nodes with pagination
	err := query.Offset(offset).
		Limit(limit).
		Order("sort ASC, created_at DESC").
		Find(&nodes).Error
	
	return nodes, total, err
}

// ListEnabled gets enabled nodes with pagination
func (r *nodeRepository) ListEnabled(offset, limit int) ([]*models.Node, int64, error) {
	var nodes []*models.Node
	var total int64
	
	query := r.db.Model(&models.Node{}).Where("is_enabled = ?", true)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get nodes with pagination
	err := query.Offset(offset).
		Limit(limit).
		Order("sort ASC, created_at DESC").
		Find(&nodes).Error
	
	return nodes, total, err
}

// ListAvailable gets available nodes (enabled and online) with pagination
func (r *nodeRepository) ListAvailable(offset, limit int) ([]*models.Node, int64, error) {
	var nodes []*models.Node
	var total int64
	
	query := r.db.Model(&models.Node{}).Where(
		"is_enabled = ? AND status = ?",
		true, models.NodeStatusOnline,
	)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get nodes with pagination
	err := query.Offset(offset).
		Limit(limit).
		Order("sort ASC, created_at DESC").
		Find(&nodes).Error
	
	return nodes, total, err
}

// Search searches nodes by name, description, or region
func (r *nodeRepository) Search(query string, offset, limit int) ([]*models.Node, int64, error) {
	var nodes []*models.Node
	var total int64
	
	searchQuery := "%" + query + "%"
	dbQuery := r.db.Model(&models.Node{}).Where(
		"name LIKE ? OR description LIKE ? OR region LIKE ? OR country LIKE ? OR city LIKE ?",
		searchQuery, searchQuery, searchQuery, searchQuery, searchQuery,
	)
	
	// Get total count
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get nodes with pagination
	err := dbQuery.Offset(offset).
		Limit(limit).
		Order("sort ASC, created_at DESC").
		Find(&nodes).Error
	
	return nodes, total, err
}

// UpdateHeartbeat updates node heartbeat timestamp
func (r *nodeRepository) UpdateHeartbeat(nodeID uint) error {
	now := time.Now()
	return r.db.Model(&models.Node{}).
		Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"last_heartbeat": now,
			"status":         models.NodeStatusOnline,
		}).Error
}

// UpdateStatus updates node status
func (r *nodeRepository) UpdateStatus(nodeID uint, status models.NodeStatus) error {
	return r.db.Model(&models.Node{}).
		Where("id = ?", nodeID).
		Update("status", status).
		Error
}

// UpdateSystemInfo updates node system information
func (r *nodeRepository) UpdateSystemInfo(nodeID uint, cpu, memory, disk, load1, load5, load15 float64) error {
	return r.db.Model(&models.Node{}).
		Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"cpu_usage":    cpu,
			"memory_usage": memory,
			"disk_usage":   disk,
			"load1":        load1,
			"load5":        load5,
			"load15":       load15,
		}).Error
}

// UpdateTraffic updates node traffic statistics
func (r *nodeRepository) UpdateTraffic(nodeID uint, upload, download int64) error {
	return r.db.Model(&models.Node{}).
		Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"upload_traffic":   gorm.Expr("upload_traffic + ?", upload),
			"download_traffic": gorm.Expr("download_traffic + ?", download),
			"total_traffic":    gorm.Expr("total_traffic + ?", upload+download),
		}).Error
}

// UpdateUserCount updates node current user count
func (r *nodeRepository) UpdateUserCount(nodeID uint, count int) error {
	return r.db.Model(&models.Node{}).
		Where("id = ?", nodeID).
		Update("current_users", count).
		Error
}

// IncrementUserCount increments node user count
func (r *nodeRepository) IncrementUserCount(nodeID uint) error {
	return r.db.Model(&models.Node{}).
		Where("id = ?", nodeID).
		UpdateColumn("current_users", gorm.Expr("current_users + 1")).
		Error
}

// DecrementUserCount decrements node user count
func (r *nodeRepository) DecrementUserCount(nodeID uint) error {
	return r.db.Model(&models.Node{}).
		Where("id = ? AND current_users > 0", nodeID).
		UpdateColumn("current_users", gorm.Expr("current_users - 1")).
		Error
}

// GetNodeCount gets total node count
func (r *nodeRepository) GetNodeCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Node{}).Count(&count).Error
	return count, err
}

// GetOnlineNodeCount gets online node count
func (r *nodeRepository) GetOnlineNodeCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Node{}).
		Where("status = ?", models.NodeStatusOnline).
		Count(&count).Error
	return count, err
}

// GetNodesByRegion gets node count by region
func (r *nodeRepository) GetNodesByRegion() (map[string]int64, error) {
	var results []struct {
		Region string
		Count  int64
	}
	
	err := r.db.Model(&models.Node{}).
		Select("region, COUNT(*) as count").
		Group("region").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	regionMap := make(map[string]int64)
	for _, result := range results {
		regionMap[result.Region] = result.Count
	}
	
	return regionMap, nil
}

// GetNodesByType gets node count by type
func (r *nodeRepository) GetNodesByType() (map[string]int64, error) {
	var results []struct {
		Type  string
		Count int64
	}
	
	err := r.db.Model(&models.Node{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	typeMap := make(map[string]int64)
	for _, result := range results {
		typeMap[result.Type] = result.Count
	}
	
	return typeMap, nil
}

// GetTopTrafficNodes gets nodes with highest traffic usage
func (r *nodeRepository) GetTopTrafficNodes(limit int) ([]*models.Node, error) {
	var nodes []*models.Node
	err := r.db.Order("total_traffic DESC").
		Limit(limit).
		Find(&nodes).Error
	return nodes, err
}

// GetNodesWithHighLoad gets nodes with high CPU or memory usage
func (r *nodeRepository) GetNodesWithHighLoad(cpuThreshold, memoryThreshold float64) ([]*models.Node, error) {
	var nodes []*models.Node
	err := r.db.Where("cpu_usage > ? OR memory_usage > ?", cpuThreshold, memoryThreshold).
		Find(&nodes).Error
	return nodes, err
}

// GetOfflineNodes gets nodes that haven't sent heartbeat within threshold
func (r *nodeRepository) GetOfflineNodes(threshold time.Duration) ([]*models.Node, error) {
	var nodes []*models.Node
	cutoff := time.Now().Add(-threshold)
	err := r.db.Where("last_heartbeat < ? OR last_heartbeat IS NULL", cutoff).
		Find(&nodes).Error
	return nodes, err
}

// GetUserNodes gets nodes accessible by a user
func (r *nodeRepository) GetUserNodes(userID uint) ([]*models.Node, error) {
	var nodes []*models.Node
	err := r.db.Table("nodes").
		Joins("JOIN user_nodes ON nodes.id = user_nodes.node_id").
		Where("user_nodes.user_id = ? AND user_nodes.is_enabled = ?", userID, true).
		Order("user_nodes.priority ASC, nodes.sort ASC").
		Find(&nodes).Error
	return nodes, err
}

// GetNodeUsers gets users who have access to a node
func (r *nodeRepository) GetNodeUsers(nodeID uint) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Table("users").
		Joins("JOIN user_nodes ON users.id = user_nodes.user_id").
		Where("user_nodes.node_id = ? AND user_nodes.is_enabled = ?", nodeID, true).
		Find(&users).Error
	return users, err
}

// AddUserToNode adds user access to a node
func (r *nodeRepository) AddUserToNode(userID, nodeID uint) error {
	userNode := &models.UserNode{
		UserID:    userID,
		NodeID:    nodeID,
		IsEnabled: true,
		Priority:  0,
	}
	return r.db.Create(userNode).Error
}

// RemoveUserFromNode removes user access from a node
func (r *nodeRepository) RemoveUserFromNode(userID, nodeID uint) error {
	return r.db.Where("user_id = ? AND node_id = ?", userID, nodeID).
		Delete(&models.UserNode{}).Error
}

// SetUserNodePriority sets priority for user-node relationship
func (r *nodeRepository) SetUserNodePriority(userID, nodeID uint, priority int) error {
	return r.db.Model(&models.UserNode{}).
		Where("user_id = ? AND node_id = ?", userID, nodeID).
		Update("priority", priority).
		Error
}

// EnableUserNode enables user access to a node
func (r *nodeRepository) EnableUserNode(userID, nodeID uint) error {
	return r.db.Model(&models.UserNode{}).
		Where("user_id = ? AND node_id = ?", userID, nodeID).
		Update("is_enabled", true).
		Error
}

// DisableUserNode disables user access to a node
func (r *nodeRepository) DisableUserNode(userID, nodeID uint) error {
	return r.db.Model(&models.UserNode{}).
		Where("user_id = ? AND node_id = ?", userID, nodeID).
		Update("is_enabled", false).
		Error
}

// BatchUpdateStatus updates status for multiple nodes
func (r *nodeRepository) BatchUpdateStatus(nodeIDs []uint, status models.NodeStatus) error {
	return r.db.Model(&models.Node{}).
		Where("id IN ?", nodeIDs).
		Update("status", status).
		Error
}

// BatchEnable enables multiple nodes
func (r *nodeRepository) BatchEnable(nodeIDs []uint) error {
	return r.db.Model(&models.Node{}).
		Where("id IN ?", nodeIDs).
		Update("is_enabled", true).
		Error
}

// BatchDisable disables multiple nodes
func (r *nodeRepository) BatchDisable(nodeIDs []uint) error {
	return r.db.Model(&models.Node{}).
		Where("id IN ?", nodeIDs).
		Update("is_enabled", false).
		Error
}

// BatchDelete soft deletes multiple nodes
func (r *nodeRepository) BatchDelete(nodeIDs []uint) error {
	return r.db.Delete(&models.Node{}, nodeIDs).Error
}

// GetNodeStats gets node statistics
func (r *nodeRepository) GetNodeStats() (*models.NodeStats, error) {
	var stats models.NodeStats
	
	// Get total nodes
	err := r.db.Model(&models.Node{}).Count(&stats.TotalNodes).Error
	if err != nil {
		return nil, err
	}
	
	// Get online nodes
	err = r.db.Model(&models.Node{}).
		Where("status = ?", models.NodeStatusOnline).
		Count(&stats.OnlineNodes).Error
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}