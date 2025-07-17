package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"sing-box-web/pkg/models"
)

// TrafficRepository interface defines traffic data access methods
type TrafficRepository interface {
	// Basic CRUD operations
	CreateRecord(record *models.TrafficRecord) error
	GetRecordByID(id uint) (*models.TrafficRecord, error)
	UpdateRecord(record *models.TrafficRecord) error
	DeleteRecord(id uint) error
	
	// List operations
	ListRecords(userID, nodeID uint, start, end time.Time, offset, limit int) ([]*models.TrafficRecord, int64, error)
	ListUserRecords(userID uint, start, end time.Time, offset, limit int) ([]*models.TrafficRecord, int64, error)
	ListNodeRecords(nodeID uint, start, end time.Time, offset, limit int) ([]*models.TrafficRecord, int64, error)
	ListRecentRecords(limit int) ([]*models.TrafficRecord, error)
	
	// Statistics operations
	GetUserTrafficSum(userID uint, start, end time.Time) (upload, download, total int64, err error)
	GetNodeTrafficSum(nodeID uint, start, end time.Time) (upload, download, total int64, err error)
	GetTotalTrafficSum(start, end time.Time) (upload, download, total int64, err error)
	GetUserDailyTraffic(userID uint, days int) ([]models.TrafficSummary, error)
	GetNodeDailyTraffic(nodeID uint, days int) ([]models.TrafficSummary, error)
	GetTopTrafficUsers(start, end time.Time, limit int) ([]*models.User, error)
	GetTopTrafficNodes(start, end time.Time, limit int) ([]*models.Node, error)
	
	// Hourly statistics
	GetHourlyTraffic(start, end time.Time) ([]models.TrafficSummary, error)
	GetUserHourlyTraffic(userID uint, start, end time.Time) ([]models.TrafficSummary, error)
	GetNodeHourlyTraffic(nodeID uint, start, end time.Time) ([]models.TrafficSummary, error)
	
	// Summary operations
	CreateSummary(summary *models.TrafficSummary) error
	GetSummaryByKey(userID, nodeID uint, date time.Time, summaryType string) (*models.TrafficSummary, error)
	UpdateSummary(summary *models.TrafficSummary) error
	UpsertSummary(summary *models.TrafficSummary) error
	ListSummaries(start, end time.Time, summaryType string, offset, limit int) ([]*models.TrafficSummary, int64, error)
	
	// Data aggregation
	AggregateHourlyData(date time.Time) error
	AggregateDailyData(date time.Time) error
	AggregateMonthlyData(date time.Time) error
	
	// Data cleanup
	CleanupOldRecords(retentionDays int) error
	CleanupOldSummaries(retentionDays int) error
	
	// Real-time operations
	GetActiveConnections() ([]*models.TrafficRecord, error)
	GetActiveUserConnections(userID uint) ([]*models.TrafficRecord, error)
	GetActiveNodeConnections(nodeID uint) ([]*models.TrafficRecord, error)
	CloseConnection(sessionID string) error
	
	// Batch operations
	BatchCreateRecords(records []*models.TrafficRecord) error
	BatchUpdateRecords(records []*models.TrafficRecord) error
	
	// Additional methods for gRPC service
	GetUserTraffic(userID uint, start, end time.Time) ([]*models.TrafficRecord, error)
	GetNodeTraffic(nodeID uint, start, end time.Time) ([]*models.TrafficRecord, error)
	GetTotalTrafficInRange(start, end time.Time) (int64, error)
}

// trafficRepository implements TrafficRepository interface
type trafficRepository struct {
	db *gorm.DB
}

// NewTrafficRepository creates a new traffic repository
func NewTrafficRepository(db *gorm.DB) TrafficRepository {
	return &trafficRepository{db: db}
}

// CreateRecord creates a new traffic record
func (r *trafficRepository) CreateRecord(record *models.TrafficRecord) error {
	return r.db.Create(record).Error
}

// GetRecordByID gets traffic record by ID
func (r *trafficRepository) GetRecordByID(id uint) (*models.TrafficRecord, error) {
	var record models.TrafficRecord
	err := r.db.Preload("User").Preload("Node").First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateRecord updates traffic record
func (r *trafficRepository) UpdateRecord(record *models.TrafficRecord) error {
	return r.db.Save(record).Error
}

// DeleteRecord soft deletes a traffic record
func (r *trafficRepository) DeleteRecord(id uint) error {
	return r.db.Delete(&models.TrafficRecord{}, id).Error
}

// ListRecords gets traffic records with filters and pagination
func (r *trafficRepository) ListRecords(userID, nodeID uint, start, end time.Time, offset, limit int) ([]*models.TrafficRecord, int64, error) {
	var records []*models.TrafficRecord
	var total int64
	
	query := r.db.Model(&models.TrafficRecord{})
	
	// Apply filters
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if nodeID > 0 {
		query = query.Where("node_id = ?", nodeID)
	}
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("record_date BETWEEN ? AND ?", start, end)
	}
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get records with pagination
	err := query.Preload("User").Preload("Node").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&records).Error
	
	return records, total, err
}

// ListUserRecords gets traffic records for a specific user
func (r *trafficRepository) ListUserRecords(userID uint, start, end time.Time, offset, limit int) ([]*models.TrafficRecord, int64, error) {
	return r.ListRecords(userID, 0, start, end, offset, limit)
}

// ListNodeRecords gets traffic records for a specific node
func (r *trafficRepository) ListNodeRecords(nodeID uint, start, end time.Time, offset, limit int) ([]*models.TrafficRecord, int64, error) {
	return r.ListRecords(0, nodeID, start, end, offset, limit)
}

// ListRecentRecords gets recent traffic records
func (r *trafficRepository) ListRecentRecords(limit int) ([]*models.TrafficRecord, error) {
	var records []*models.TrafficRecord
	err := r.db.Preload("User").Preload("Node").
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// GetUserTrafficSum gets total traffic for a user within date range
func (r *trafficRepository) GetUserTrafficSum(userID uint, start, end time.Time) (upload, download, total int64, err error) {
	var result struct {
		Upload   int64
		Download int64
		Total    int64
	}
	
	query := r.db.Model(&models.TrafficRecord{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download, COALESCE(SUM(total), 0) as total").
		Where("user_id = ?", userID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("record_date BETWEEN ? AND ?", start, end)
	}
	
	err = query.Scan(&result).Error
	return result.Upload, result.Download, result.Total, err
}

// GetNodeTrafficSum gets total traffic for a node within date range
func (r *trafficRepository) GetNodeTrafficSum(nodeID uint, start, end time.Time) (upload, download, total int64, err error) {
	var result struct {
		Upload   int64
		Download int64
		Total    int64
	}
	
	query := r.db.Model(&models.TrafficRecord{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download, COALESCE(SUM(total), 0) as total").
		Where("node_id = ?", nodeID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("record_date BETWEEN ? AND ?", start, end)
	}
	
	err = query.Scan(&result).Error
	return result.Upload, result.Download, result.Total, err
}

// GetTotalTrafficSum gets total traffic for all users within date range
func (r *trafficRepository) GetTotalTrafficSum(start, end time.Time) (upload, download, total int64, err error) {
	var result struct {
		Upload   int64
		Download int64
		Total    int64
	}
	
	query := r.db.Model(&models.TrafficRecord{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download, COALESCE(SUM(total), 0) as total")
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("record_date BETWEEN ? AND ?", start, end)
	}
	
	err = query.Scan(&result).Error
	return result.Upload, result.Download, result.Total, err
}

// GetUserDailyTraffic gets daily traffic summary for a user
func (r *trafficRepository) GetUserDailyTraffic(userID uint, days int) ([]models.TrafficSummary, error) {
	var summaries []models.TrafficSummary
	
	start := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	
	err := r.db.Where("user_id = ? AND summary_type = ? AND summary_date >= ?", 
		userID, "daily", start).
		Order("summary_date DESC").
		Find(&summaries).Error
	
	return summaries, err
}

// GetNodeDailyTraffic gets daily traffic summary for a node
func (r *trafficRepository) GetNodeDailyTraffic(nodeID uint, days int) ([]models.TrafficSummary, error) {
	var summaries []models.TrafficSummary
	
	start := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	
	err := r.db.Where("node_id = ? AND summary_type = ? AND summary_date >= ?", 
		nodeID, "daily", start).
		Order("summary_date DESC").
		Find(&summaries).Error
	
	return summaries, err
}

// GetTopTrafficUsers gets users with highest traffic usage
func (r *trafficRepository) GetTopTrafficUsers(start, end time.Time, limit int) ([]*models.User, error) {
	var users []*models.User
	
	subQuery := r.db.Model(&models.TrafficRecord{}).
		Select("user_id, SUM(total) as total_traffic").
		Where("record_date BETWEEN ? AND ?", start, end).
		Group("user_id").
		Order("total_traffic DESC").
		Limit(limit)
	
	err := r.db.Table("users").
		Joins("JOIN (?) as traffic_stats ON users.id = traffic_stats.user_id", subQuery).
		Preload("Plan").
		Find(&users).Error
	
	return users, err
}

// GetTopTrafficNodes gets nodes with highest traffic usage
func (r *trafficRepository) GetTopTrafficNodes(start, end time.Time, limit int) ([]*models.Node, error) {
	var nodes []*models.Node
	
	subQuery := r.db.Model(&models.TrafficRecord{}).
		Select("node_id, SUM(total) as total_traffic").
		Where("record_date BETWEEN ? AND ?", start, end).
		Group("node_id").
		Order("total_traffic DESC").
		Limit(limit)
	
	err := r.db.Table("nodes").
		Joins("JOIN (?) as traffic_stats ON nodes.id = traffic_stats.node_id", subQuery).
		Find(&nodes).Error
	
	return nodes, err
}

// GetHourlyTraffic gets hourly traffic statistics
func (r *trafficRepository) GetHourlyTraffic(start, end time.Time) ([]models.TrafficSummary, error) {
	var summaries []models.TrafficSummary
	
	err := r.db.Model(&models.TrafficRecord{}).
		Select(`
			DATE(record_date) as summary_date,
			record_hour,
			SUM(upload) as total_upload,
			SUM(download) as total_download,
			SUM(total) as total_traffic,
			COUNT(*) as total_connections
		`).
		Where("record_date BETWEEN ? AND ?", start, end).
		Group("DATE(record_date), record_hour").
		Order("summary_date DESC, record_hour DESC").
		Scan(&summaries).Error
	
	return summaries, err
}

// GetUserHourlyTraffic gets hourly traffic for a specific user
func (r *trafficRepository) GetUserHourlyTraffic(userID uint, start, end time.Time) ([]models.TrafficSummary, error) {
	var summaries []models.TrafficSummary
	
	err := r.db.Model(&models.TrafficRecord{}).
		Select(`
			user_id,
			DATE(record_date) as summary_date,
			record_hour,
			SUM(upload) as total_upload,
			SUM(download) as total_download,
			SUM(total) as total_traffic,
			COUNT(*) as total_connections
		`).
		Where("user_id = ? AND record_date BETWEEN ? AND ?", userID, start, end).
		Group("user_id, DATE(record_date), record_hour").
		Order("summary_date DESC, record_hour DESC").
		Scan(&summaries).Error
	
	return summaries, err
}

// GetNodeHourlyTraffic gets hourly traffic for a specific node
func (r *trafficRepository) GetNodeHourlyTraffic(nodeID uint, start, end time.Time) ([]models.TrafficSummary, error) {
	var summaries []models.TrafficSummary
	
	err := r.db.Model(&models.TrafficRecord{}).
		Select(`
			node_id,
			DATE(record_date) as summary_date,
			record_hour,
			SUM(upload) as total_upload,
			SUM(download) as total_download,
			SUM(total) as total_traffic,
			COUNT(*) as total_connections
		`).
		Where("node_id = ? AND record_date BETWEEN ? AND ?", nodeID, start, end).
		Group("node_id, DATE(record_date), record_hour").
		Order("summary_date DESC, record_hour DESC").
		Scan(&summaries).Error
	
	return summaries, err
}

// CreateSummary creates a new traffic summary
func (r *trafficRepository) CreateSummary(summary *models.TrafficSummary) error {
	return r.db.Create(summary).Error
}

// GetSummaryByKey gets traffic summary by key fields
func (r *trafficRepository) GetSummaryByKey(userID, nodeID uint, date time.Time, summaryType string) (*models.TrafficSummary, error) {
	var summary models.TrafficSummary
	err := r.db.Where("user_id = ? AND node_id = ? AND summary_date = ? AND summary_type = ?", 
		userID, nodeID, date, summaryType).First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// UpdateSummary updates traffic summary
func (r *trafficRepository) UpdateSummary(summary *models.TrafficSummary) error {
	return r.db.Save(summary).Error
}

// UpsertSummary creates or updates traffic summary
func (r *trafficRepository) UpsertSummary(summary *models.TrafficSummary) error {
	return r.db.Where("user_id = ? AND node_id = ? AND summary_date = ? AND summary_type = ?",
		summary.UserID, summary.NodeID, summary.SummaryDate, summary.SummaryType).
		Assign(summary).
		FirstOrCreate(&summary).Error
}

// ListSummaries gets traffic summaries with pagination
func (r *trafficRepository) ListSummaries(start, end time.Time, summaryType string, offset, limit int) ([]*models.TrafficSummary, int64, error) {
	var summaries []*models.TrafficSummary
	var total int64
	
	query := r.db.Model(&models.TrafficSummary{}).
		Where("summary_date BETWEEN ? AND ? AND summary_type = ?", start, end, summaryType)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get summaries with pagination
	err := query.Preload("User").Preload("Node").
		Offset(offset).
		Limit(limit).
		Order("summary_date DESC").
		Find(&summaries).Error
	
	return summaries, total, err
}

// AggregateHourlyData aggregates traffic data into hourly summaries
func (r *trafficRepository) AggregateHourlyData(date time.Time) error {
	// This would typically be implemented as a SQL procedure or complex query
	// For now, we'll use a simplified Go implementation
	return r.db.Transaction(func(tx *gorm.DB) error {
		var records []models.TrafficRecord
		
		// Get all records for the date
		if err := tx.Where("record_date = ?", date.Truncate(24*time.Hour)).Find(&records).Error; err != nil {
			return err
		}
		
		// Group by user, node, and hour
		summaryMap := make(map[string]*models.TrafficSummary)
		
		for _, record := range records {
			key := fmt.Sprintf("%d-%d-%d", record.UserID, record.NodeID, record.RecordHour)
			
			if summary, exists := summaryMap[key]; exists {
				summary.TotalUpload += record.Upload
				summary.TotalDownload += record.Download
				summary.TotalTraffic += record.Total
				summary.TotalConnections++
			} else {
				summaryMap[key] = &models.TrafficSummary{
					UserID:           record.UserID,
					NodeID:           record.NodeID,
					SummaryDate:      date.Truncate(24 * time.Hour),
					SummaryType:      "hourly",
					TotalUpload:      record.Upload,
					TotalDownload:    record.Download,
					TotalTraffic:     record.Total,
					TotalConnections: 1,
				}
			}
		}
		
		// Save summaries
		for _, summary := range summaryMap {
			if err := r.UpsertSummary(summary); err != nil {
				return err
			}
		}
		
		return nil
	})
}

// AggregateDailyData aggregates traffic data into daily summaries
func (r *trafficRepository) AggregateDailyData(date time.Time) error {
	return r.db.Exec(`
		INSERT INTO traffic_summaries (user_id, node_id, summary_date, summary_type, 
			total_upload, total_download, total_traffic, total_connections, created_at, updated_at)
		SELECT user_id, node_id, ?, 'daily',
			SUM(upload), SUM(download), SUM(total), COUNT(*), NOW(), NOW()
		FROM traffic_records 
		WHERE record_date = ?
		GROUP BY user_id, node_id
		ON DUPLICATE KEY UPDATE
			total_upload = VALUES(total_upload),
			total_download = VALUES(total_download),
			total_traffic = VALUES(total_traffic),
			total_connections = VALUES(total_connections),
			updated_at = NOW()
	`, date.Truncate(24*time.Hour), date.Truncate(24*time.Hour)).Error
}

// AggregateMonthlyData aggregates traffic data into monthly summaries
func (r *trafficRepository) AggregateMonthlyData(date time.Time) error {
	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	
	return r.db.Exec(`
		INSERT INTO traffic_summaries (user_id, node_id, summary_date, summary_type, 
			total_upload, total_download, total_traffic, total_connections, created_at, updated_at)
		SELECT user_id, node_id, ?, 'monthly',
			SUM(upload), SUM(download), SUM(total), COUNT(*), NOW(), NOW()
		FROM traffic_records 
		WHERE record_date >= ? AND record_date < ?
		GROUP BY user_id, node_id
		ON DUPLICATE KEY UPDATE
			total_upload = VALUES(total_upload),
			total_download = VALUES(total_download),
			total_traffic = VALUES(total_traffic),
			total_connections = VALUES(total_connections),
			updated_at = NOW()
	`, monthStart, monthStart.AddDate(0, 1, 0), monthStart).Error
}

// CleanupOldRecords removes old traffic records
func (r *trafficRepository) CleanupOldRecords(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return r.db.Where("created_at < ?", cutoff).Delete(&models.TrafficRecord{}).Error
}

// CleanupOldSummaries removes old traffic summaries
func (r *trafficRepository) CleanupOldSummaries(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return r.db.Where("summary_date < ?", cutoff).Delete(&models.TrafficSummary{}).Error
}

// GetActiveConnections gets all active connections
func (r *trafficRepository) GetActiveConnections() ([]*models.TrafficRecord, error) {
	var records []*models.TrafficRecord
	err := r.db.Preload("User").Preload("Node").
		Where("disconnect_time IS NULL").
		Order("connect_time DESC").
		Find(&records).Error
	return records, err
}

// GetActiveUserConnections gets active connections for a specific user
func (r *trafficRepository) GetActiveUserConnections(userID uint) ([]*models.TrafficRecord, error) {
	var records []*models.TrafficRecord
	err := r.db.Preload("User").Preload("Node").
		Where("user_id = ? AND disconnect_time IS NULL", userID).
		Order("connect_time DESC").
		Find(&records).Error
	return records, err
}

// GetActiveNodeConnections gets active connections for a specific node
func (r *trafficRepository) GetActiveNodeConnections(nodeID uint) ([]*models.TrafficRecord, error) {
	var records []*models.TrafficRecord
	err := r.db.Preload("User").Preload("Node").
		Where("node_id = ? AND disconnect_time IS NULL", nodeID).
		Order("connect_time DESC").
		Find(&records).Error
	return records, err
}

// CloseConnection closes an active connection
func (r *trafficRepository) CloseConnection(sessionID string) error {
	now := time.Now()
	return r.db.Model(&models.TrafficRecord{}).
		Where("session_id = ? AND disconnect_time IS NULL", sessionID).
		Updates(map[string]interface{}{
			"disconnect_time": now,
			"duration":        gorm.Expr("TIMESTAMPDIFF(SECOND, connect_time, ?)", now),
		}).Error
}

// BatchCreateRecords creates multiple traffic records
func (r *trafficRepository) BatchCreateRecords(records []*models.TrafficRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.CreateInBatches(records, 100).Error
}

// BatchUpdateRecords updates multiple traffic records
func (r *trafficRepository) BatchUpdateRecords(records []*models.TrafficRecord) error {
	if len(records) == 0 {
		return nil
	}
	
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, record := range records {
			if err := tx.Save(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetUserTraffic gets traffic records for a specific user in time range
func (r *trafficRepository) GetUserTraffic(userID uint, start, end time.Time) ([]*models.TrafficRecord, error) {
	var records []*models.TrafficRecord
	query := r.db.Preload("User").Preload("Node").Where("user_id = ?", userID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("record_date BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("record_date DESC").Find(&records).Error
	return records, err
}

// GetNodeTraffic gets traffic records for a specific node in time range
func (r *trafficRepository) GetNodeTraffic(nodeID uint, start, end time.Time) ([]*models.TrafficRecord, error) {
	var records []*models.TrafficRecord
	query := r.db.Preload("User").Preload("Node").Where("node_id = ?", nodeID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("record_date BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("record_date DESC").Find(&records).Error
	return records, err
}

// GetTotalTrafficInRange gets total traffic in a time range
func (r *trafficRepository) GetTotalTrafficInRange(start, end time.Time) (int64, error) {
	var total int64
	query := r.db.Model(&models.TrafficRecord{}).Select("COALESCE(SUM(total), 0)")
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("record_date BETWEEN ? AND ?", start, end)
	}
	
	err := query.Scan(&total).Error
	return total, err
}