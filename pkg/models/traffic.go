package models

import (
	"time"

	"gorm.io/gorm"
)

// TrafficRecord represents traffic usage records
type TrafficRecord struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Foreign keys
	UserID uint `json:"user_id" gorm:"not null;index"`
	NodeID uint `json:"node_id" gorm:"not null;index"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Node Node `json:"node,omitempty" gorm:"foreignKey:NodeID"`

	// Traffic data
	Upload   int64 `json:"upload" gorm:"not null;default:0;comment:Upload bytes"`
	Download int64 `json:"download" gorm:"not null;default:0;comment:Download bytes"`
	Total    int64 `json:"total" gorm:"not null;default:0;comment:Total bytes"`

	// Time period
	RecordDate time.Time `json:"record_date" gorm:"not null;index;comment:Date of the record"`
	RecordHour int       `json:"record_hour" gorm:"not null;index;comment:Hour of the record (0-23)"`

	// Session information
	SessionID    string    `json:"session_id" gorm:"size:64;index;comment:Session identifier"`
	ConnectTime  time.Time `json:"connect_time" gorm:"comment:Connection start time"`
	DisconnectTime *time.Time `json:"disconnect_time,omitempty" gorm:"comment:Connection end time"`
	Duration     int64     `json:"duration" gorm:"not null;default:0;comment:Connection duration in seconds"`

	// Client information
	ClientIP    string `json:"client_ip" gorm:"size:45;comment:Client IP address"`
	UserAgent   string `json:"user_agent" gorm:"size:512;comment:User agent"`
	DeviceID    string `json:"device_id" gorm:"size:128;comment:Device identifier"`
	Protocol    string `json:"protocol" gorm:"size:32;comment:Connection protocol"`

	// Quality metrics
	AvgSpeed     int64   `json:"avg_speed" gorm:"not null;default:0;comment:Average speed in bytes/sec"`
	MaxSpeed     int64   `json:"max_speed" gorm:"not null;default:0;comment:Maximum speed in bytes/sec"`
	PacketLoss   float64 `json:"packet_loss" gorm:"type:decimal(5,2);default:0;comment:Packet loss percentage"`
	Latency      int     `json:"latency" gorm:"not null;default:0;comment:Average latency in ms"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" gorm:"serializer:json"`
}

// TableName returns the table name for TrafficRecord model
func (TrafficRecord) TableName() string {
	return "traffic_records"
}

// BeforeCreate GORM hook to calculate total before creating
func (tr *TrafficRecord) BeforeCreate(tx *gorm.DB) error {
	tr.Total = tr.Upload + tr.Download
	if tr.RecordDate.IsZero() {
		tr.RecordDate = time.Now().Truncate(24 * time.Hour)
	}
	if tr.RecordHour == 0 {
		tr.RecordHour = time.Now().Hour()
	}
	return nil
}

// BeforeUpdate GORM hook to calculate total before updating
func (tr *TrafficRecord) BeforeUpdate(tx *gorm.DB) error {
	tr.Total = tr.Upload + tr.Download
	return nil
}

// CalculateDuration calculates and updates the duration
func (tr *TrafficRecord) CalculateDuration() {
	if tr.DisconnectTime != nil {
		tr.Duration = int64(tr.DisconnectTime.Sub(tr.ConnectTime).Seconds())
	} else {
		tr.Duration = int64(time.Since(tr.ConnectTime).Seconds())
	}
}

// CalculateAvgSpeed calculates average speed
func (tr *TrafficRecord) CalculateAvgSpeed() {
	if tr.Duration > 0 {
		tr.AvgSpeed = tr.Total / tr.Duration
	}
}

// TrafficSummary represents traffic summary for reporting
type TrafficSummary struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Summary key
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	NodeID     uint      `json:"node_id" gorm:"not null;index"`
	SummaryDate time.Time `json:"summary_date" gorm:"not null;index;comment:Summary date"`
	SummaryType string    `json:"summary_type" gorm:"not null;size:10;index;comment:daily/monthly/yearly"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Node Node `json:"node,omitempty" gorm:"foreignKey:NodeID"`

	// Aggregated data
	TotalUpload   int64 `json:"total_upload" gorm:"not null;default:0"`
	TotalDownload int64 `json:"total_download" gorm:"not null;default:0"`
	TotalTraffic  int64 `json:"total_traffic" gorm:"not null;default:0"`

	// Connection statistics
	TotalConnections int64 `json:"total_connections" gorm:"not null;default:0"`
	TotalDuration    int64 `json:"total_duration" gorm:"not null;default:0;comment:Total duration in seconds"`
	AvgDuration      int64 `json:"avg_duration" gorm:"not null;default:0;comment:Average duration in seconds"`

	// Performance metrics
	AvgSpeed       int64   `json:"avg_speed" gorm:"not null;default:0"`
	MaxSpeed       int64   `json:"max_speed" gorm:"not null;default:0"`
	AvgPacketLoss  float64 `json:"avg_packet_loss" gorm:"type:decimal(5,2);default:0"`
	AvgLatency     int     `json:"avg_latency" gorm:"not null;default:0"`

	// Peak usage
	PeakHour       int   `json:"peak_hour" gorm:"comment:Peak usage hour (0-23)"`
	PeakHourTraffic int64 `json:"peak_hour_traffic" gorm:"comment:Traffic during peak hour"`
}

// TableName returns the table name for TrafficSummary model
func (TrafficSummary) TableName() string {
	return "traffic_summaries"
}

// BeforeCreate GORM hook for TrafficSummary
func (ts *TrafficSummary) BeforeCreate(tx *gorm.DB) error {
	ts.TotalTraffic = ts.TotalUpload + ts.TotalDownload
	if ts.TotalConnections > 0 {
		ts.AvgDuration = ts.TotalDuration / ts.TotalConnections
	}
	return nil
}

// BeforeUpdate GORM hook for TrafficSummary
func (ts *TrafficSummary) BeforeUpdate(tx *gorm.DB) error {
	ts.TotalTraffic = ts.TotalUpload + ts.TotalDownload
	if ts.TotalConnections > 0 {
		ts.AvgDuration = ts.TotalDuration / ts.TotalConnections
	}
	return nil
}

// TrafficQuota represents traffic quota policies
type TrafficQuota struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Policy information
	Name        string `json:"name" gorm:"not null;size:128"`
	Description string `json:"description" gorm:"type:text"`

	// Quota settings
	QuotaBytes    int64  `json:"quota_bytes" gorm:"not null;comment:Quota in bytes"`
	ResetPeriod   string `json:"reset_period" gorm:"not null;size:16;comment:daily/weekly/monthly/yearly"`
	ResetDay      int    `json:"reset_day" gorm:"comment:Day of month for monthly reset (1-31)"`
	ResetWeekday  int    `json:"reset_weekday" gorm:"comment:Day of week for weekly reset (0-6)"`

	// Rate limiting
	SpeedLimitUp   int64 `json:"speed_limit_up" gorm:"default:0;comment:Upload speed limit in bytes/sec"`
	SpeedLimitDown int64 `json:"speed_limit_down" gorm:"default:0;comment:Download speed limit in bytes/sec"`

	// Behavior when quota exceeded
	ActionOnExceed string `json:"action_on_exceed" gorm:"not null;size:20;default:'block';comment:block/throttle/notify"`
	ThrottleSpeed  int64  `json:"throttle_speed" gorm:"default:0;comment:Throttle speed when exceeded"`

	// Warning settings
	WarningThreshold float64 `json:"warning_threshold" gorm:"type:decimal(3,2);default:0.8;comment:Warning threshold (0.0-1.0)"`
	NotifyOnWarning  bool    `json:"notify_on_warning" gorm:"default:true"`
	NotifyOnExceed   bool    `json:"notify_on_exceed" gorm:"default:true"`

	// Status
	IsEnabled bool `json:"is_enabled" gorm:"not null;default:true"`
	Priority  int  `json:"priority" gorm:"not null;default:0;comment:Higher number means higher priority"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" gorm:"serializer:json"`
}

// TableName returns the table name for TrafficQuota model
func (TrafficQuota) TableName() string {
	return "traffic_quotas"
}

// GetNextResetTime calculates the next reset time based on the reset period
func (tq *TrafficQuota) GetNextResetTime(from time.Time) time.Time {
	switch tq.ResetPeriod {
	case "daily":
		return from.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	case "weekly":
		// Reset on specified weekday
		days := int(tq.ResetWeekday) - int(from.Weekday())
		if days <= 0 {
			days += 7
		}
		return from.AddDate(0, 0, days).Truncate(24 * time.Hour)
	case "monthly":
		// Reset on specified day of month
		year, month, _ := from.Date()
		resetDay := tq.ResetDay
		if resetDay <= 0 {
			resetDay = 1
		}
		nextMonth := time.Date(year, month+1, resetDay, 0, 0, 0, 0, from.Location())
		// Handle cases where the day doesn't exist in the month
		if nextMonth.Day() != resetDay {
			nextMonth = time.Date(year, month+2, 1, 0, 0, 0, 0, from.Location()).AddDate(0, 0, -1)
		}
		return nextMonth
	case "yearly":
		return from.AddDate(1, 0, 0).Truncate(24 * time.Hour)
	default:
		// Default to monthly
		return from.AddDate(0, 1, 0).Truncate(24 * time.Hour)
	}
}

// IsExceeded checks if the usage exceeds the quota
func (tq *TrafficQuota) IsExceeded(usage int64) bool {
	return tq.QuotaBytes > 0 && usage >= tq.QuotaBytes
}

// IsWarning checks if the usage is in warning range
func (tq *TrafficQuota) IsWarning(usage int64) bool {
	if tq.QuotaBytes <= 0 {
		return false
	}
	threshold := float64(tq.QuotaBytes) * tq.WarningThreshold
	return float64(usage) >= threshold
}

// GetUsagePercentage returns usage percentage
func (tq *TrafficQuota) GetUsagePercentage(usage int64) float64 {
	if tq.QuotaBytes <= 0 {
		return 0
	}
	return float64(usage) / float64(tq.QuotaBytes) * 100
}