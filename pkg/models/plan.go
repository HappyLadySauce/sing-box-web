package models

import (
	"time"

	"gorm.io/gorm"
)

// PlanStatus represents plan status
type PlanStatus string

const (
	PlanStatusActive   PlanStatus = "active"
	PlanStatusInactive PlanStatus = "inactive"
	PlanStatusArchived PlanStatus = "archived"
)

// PlanPeriod represents billing period
type PlanPeriod string

const (
	PlanPeriodDaily   PlanPeriod = "daily"
	PlanPeriodWeekly  PlanPeriod = "weekly"
	PlanPeriodMonthly PlanPeriod = "monthly"
	PlanPeriodYearly  PlanPeriod = "yearly"
	PlanPeriodLifetime PlanPeriod = "lifetime"
)

// Plan represents a subscription plan
type Plan struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Basic information
	Name        string     `json:"name" gorm:"not null;size:128"`
	Description string     `json:"description" gorm:"type:text"`
	Status      PlanStatus `json:"status" gorm:"not null;default:'active';size:20"`
	
	// Billing
	Period    PlanPeriod `json:"period" gorm:"not null;size:20"`
	Price     int64      `json:"price" gorm:"not null;default:0;comment:Price in cents"`
	Currency  string     `json:"currency" gorm:"not null;default:'USD';size:3"`
	
	// Traffic limits
	TrafficQuota int64 `json:"traffic_quota" gorm:"not null;default:0;comment:Monthly traffic quota in bytes, 0 = unlimited"`
	SpeedLimit   int64 `json:"speed_limit" gorm:"not null;default:0;comment:Speed limit in bytes/sec, 0 = unlimited"`
	
	// Connection limits
	DeviceLimit      int `json:"device_limit" gorm:"not null;default:1;comment:Maximum concurrent devices"`
	ConnectionLimit  int `json:"connection_limit" gorm:"not null;default:0;comment:Maximum concurrent connections, 0 = unlimited"`
	
	// Features
	AllowedProtocols string `json:"allowed_protocols" gorm:"size:512;comment:Comma-separated list of allowed protocols"`
	AllowedNodes     string `json:"allowed_nodes" gorm:"type:text;comment:JSON array of allowed node IDs"`
	
	// Advanced features
	EnableFileSharing   bool `json:"enable_file_sharing" gorm:"not null;default:false"`
	EnablePortForwarding bool `json:"enable_port_forwarding" gorm:"not null;default:false"`
	EnableP2P          bool `json:"enable_p2p" gorm:"not null;default:false"`
	EnableTorrent      bool `json:"enable_torrent" gorm:"not null;default:false"`
	
	// Quality of Service
	Priority        int     `json:"priority" gorm:"not null;default:0;comment:Higher number means higher priority"`
	BandwidthRatio  float64 `json:"bandwidth_ratio" gorm:"type:decimal(3,2);default:1.0;comment:Bandwidth allocation ratio"`
	
	// Restrictions
	RestrictionLevel int    `json:"restriction_level" gorm:"not null;default:0;comment:0=none, 1=low, 2=medium, 3=high"`
	BlockedDomains   string `json:"blocked_domains" gorm:"type:text;comment:Comma-separated list of blocked domains"`
	AllowedCountries string `json:"allowed_countries" gorm:"size:512;comment:Comma-separated list of allowed country codes"`
	
	// Trial and promotion
	IsTrialPlan    bool `json:"is_trial_plan" gorm:"not null;default:false"`
	TrialDays      int  `json:"trial_days" gorm:"not null;default:0"`
	IsPromotional  bool `json:"is_promotional" gorm:"not null;default:false"`
	PromotionPrice int64 `json:"promotion_price" gorm:"default:0;comment:Promotional price in cents"`
	PromotionEndsAt *time.Time `json:"promotion_ends_at,omitempty"`
	
	// Availability
	IsPublic     bool       `json:"is_public" gorm:"not null;default:true;comment:Is visible to public"`
	IsEnabled    bool       `json:"is_enabled" gorm:"not null;default:true"`
	ValidFrom    *time.Time `json:"valid_from,omitempty"`
	ValidUntil   *time.Time `json:"valid_until,omitempty"`
	MaxUsers     int        `json:"max_users" gorm:"not null;default:0;comment:Maximum users for this plan, 0 = unlimited"`
	CurrentUsers int        `json:"current_users" gorm:"not null;default:0"`
	
	// Display
	Color       string `json:"color" gorm:"size:7;comment:Hex color code"`
	Icon        string `json:"icon" gorm:"size:64;comment:Icon identifier"`
	SortOrder   int    `json:"sort_order" gorm:"not null;default:0"`
	IsRecommended bool `json:"is_recommended" gorm:"not null;default:false"`
	
	// Metadata
	Features map[string]interface{} `json:"features,omitempty" gorm:"serializer:json;comment:Additional features"`
	Metadata map[string]interface{} `json:"metadata,omitempty" gorm:"serializer:json"`
	
	// Relationships
	Users []User `json:"users,omitempty" gorm:"foreignKey:PlanID"`
}

// TableName returns the table name for Plan model
func (Plan) TableName() string {
	return "plans"
}

// IsActive checks if plan is currently active
func (p *Plan) IsActive() bool {
	if !p.IsEnabled || p.Status != PlanStatusActive {
		return false
	}
	
	now := time.Now()
	if p.ValidFrom != nil && now.Before(*p.ValidFrom) {
		return false
	}
	if p.ValidUntil != nil && now.After(*p.ValidUntil) {
		return false
	}
	
	return true
}

// IsAvailable checks if plan is available for new users
func (p *Plan) IsAvailable() bool {
	if !p.IsActive() || !p.IsPublic {
		return false
	}
	
	if p.MaxUsers > 0 && p.CurrentUsers >= p.MaxUsers {
		return false
	}
	
	return true
}

// GetCurrentPrice returns the current effective price
func (p *Plan) GetCurrentPrice() int64 {
	if p.IsPromotional && p.PromotionEndsAt != nil && time.Now().Before(*p.PromotionEndsAt) {
		return p.PromotionPrice
	}
	return p.Price
}

// GetTrafficQuotaGB returns traffic quota in GB
func (p *Plan) GetTrafficQuotaGB() float64 {
	if p.TrafficQuota <= 0 {
		return -1 // Unlimited
	}
	return float64(p.TrafficQuota) / (1024 * 1024 * 1024)
}

// GetSpeedLimitMbps returns speed limit in Mbps
func (p *Plan) GetSpeedLimitMbps() float64 {
	if p.SpeedLimit <= 0 {
		return -1 // Unlimited
	}
	return float64(p.SpeedLimit) * 8 / (1024 * 1024)
}

// CanAcceptNewUser checks if plan can accept new users
func (p *Plan) CanAcceptNewUser() bool {
	return p.IsAvailable()
}

// GetUsagePercentage returns user usage percentage
func (p *Plan) GetUsagePercentage() float64 {
	if p.MaxUsers <= 0 {
		return 0 // Unlimited
	}
	return float64(p.CurrentUsers) / float64(p.MaxUsers) * 100
}

// IncrementUsers increments the current user count
func (p *Plan) IncrementUsers() {
	p.CurrentUsers++
}

// DecrementUsers decrements the current user count
func (p *Plan) DecrementUsers() {
	if p.CurrentUsers > 0 {
		p.CurrentUsers--
	}
}

// PlanFeature represents individual plan features
type PlanFeature struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	PlanID uint `json:"plan_id" gorm:"not null;index"`
	Plan   Plan `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	
	// Feature details
	Name        string `json:"name" gorm:"not null;size:128"`
	Description string `json:"description" gorm:"type:text"`
	Type        string `json:"type" gorm:"not null;size:32;comment:boolean/numeric/string/json"`
	Value       string `json:"value" gorm:"type:text;comment:Feature value"`
	
	// Display
	Icon      string `json:"icon" gorm:"size:64"`
	SortOrder int    `json:"sort_order" gorm:"not null;default:0"`
	IsVisible bool   `json:"is_visible" gorm:"not null;default:true"`
}

// TableName returns the table name for PlanFeature model
func (PlanFeature) TableName() string {
	return "plan_features"
}

// PlanNodeAccess represents which nodes are accessible by a plan
type PlanNodeAccess struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	PlanID uint `json:"plan_id" gorm:"not null;index"`
	NodeID uint `json:"node_id" gorm:"not null;index"`
	
	// Relationships
	Plan Plan `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	Node Node `json:"node,omitempty" gorm:"foreignKey:NodeID"`
	
	// Access control
	IsEnabled bool `json:"is_enabled" gorm:"not null;default:true"`
	Priority  int  `json:"priority" gorm:"not null;default:0;comment:Lower number means higher priority"`
	
	// Limits specific to this plan-node combination
	SpeedLimitOverride int64 `json:"speed_limit_override" gorm:"default:0;comment:Override plan speed limit for this node"`
	MaxConnections     int   `json:"max_connections" gorm:"default:0;comment:Maximum connections to this node"`
}

// TableName returns the table name for PlanNodeAccess model
func (PlanNodeAccess) TableName() string {
	return "plan_node_access"
}