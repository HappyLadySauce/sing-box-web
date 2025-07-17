package models

import (
	"time"

	"gorm.io/gorm"
)

// UserStatus represents user account status
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusExpired   UserStatus = "expired"
	UserStatusDisabled  UserStatus = "disabled"
)

// User represents a sing-box user
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Basic information
	Username    string     `json:"username" gorm:"uniqueIndex;not null;size:64"`
	Email       string     `json:"email" gorm:"uniqueIndex;size:255"`
	Password    string     `json:"-" gorm:"not null;size:255"` // Exclude from JSON
	DisplayName string     `json:"display_name" gorm:"size:128"`
	Avatar      string     `json:"avatar" gorm:"size:512"`
	Status      UserStatus `json:"status" gorm:"not null;default:'active';size:20"`

	// Plan and quota
	PlanID            uint      `json:"plan_id" gorm:"not null"`
	Plan              Plan      `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	TrafficQuota      int64     `json:"traffic_quota" gorm:"not null;default:0;comment:Monthly traffic quota in bytes"`
	TrafficUsed       int64     `json:"traffic_used" gorm:"not null;default:0;comment:Used traffic in current period"`
	TrafficResetDate  time.Time `json:"traffic_reset_date" gorm:"comment:Next traffic reset date"`
	DeviceLimit       int       `json:"device_limit" gorm:"not null;default:1;comment:Maximum concurrent devices"`
	SpeedLimit        int64     `json:"speed_limit" gorm:"not null;default:0;comment:Speed limit in bytes/sec"`

	// Account validity
	ExpiresAt    *time.Time `json:"expires_at,omitempty" gorm:"comment:Account expiration time"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP  string     `json:"last_login_ip" gorm:"size:45"`
	LoginAttempts int       `json:"login_attempts" gorm:"not null;default:0"`
	LockedUntil  *time.Time `json:"locked_until,omitempty"`

	// Subscription and configuration
	UUID         string `json:"uuid" gorm:"uniqueIndex;not null;size:36;comment:User UUID for sing-box config"`
	SubscriptionToken string `json:"subscription_token" gorm:"uniqueIndex;size:64;comment:Subscription token"`
	ConfigVersion     int    `json:"config_version" gorm:"not null;default:0;comment:Configuration version"`

	// Metadata
	Notes    string            `json:"notes" gorm:"type:text;comment:Admin notes"`
	Metadata map[string]string `json:"metadata,omitempty" gorm:"serializer:json;comment:Additional metadata"`

	// Relationships
	TrafficRecords []TrafficRecord `json:"traffic_records,omitempty" gorm:"foreignKey:UserID"`
	UserNodes      []UserNode      `json:"user_nodes,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for User model
func (User) TableName() string {
	return "users"
}

// IsActive checks if user account is active
func (u *User) IsActive() bool {
	if u.Status != UserStatusActive {
		return false
	}
	if u.ExpiresAt != nil && u.ExpiresAt.Before(time.Now()) {
		return false
	}
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return false
	}
	return true
}

// IsTrafficExceeded checks if user has exceeded traffic quota
func (u *User) IsTrafficExceeded() bool {
	return u.TrafficQuota > 0 && u.TrafficUsed >= u.TrafficQuota
}

// RemainingTraffic returns remaining traffic in bytes
func (u *User) RemainingTraffic() int64 {
	if u.TrafficQuota <= 0 {
		return -1 // Unlimited
	}
	remaining := u.TrafficQuota - u.TrafficUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ShouldResetTraffic checks if traffic should be reset
func (u *User) ShouldResetTraffic() bool {
	return !u.TrafficResetDate.IsZero() && time.Now().After(u.TrafficResetDate)
}

// ResetTraffic resets user traffic and sets next reset date
func (u *User) ResetTraffic() {
	u.TrafficUsed = 0
	u.TrafficResetDate = time.Now().AddDate(0, 1, 0) // Next month
}

// BeforeCreate GORM hook to set defaults before creating
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		// Generate UUID if not set
		u.UUID = generateUUID()
	}
	if u.SubscriptionToken == "" {
		// Generate subscription token if not set
		u.SubscriptionToken = generateToken(32)
	}
	if u.TrafficResetDate.IsZero() {
		u.TrafficResetDate = time.Now().AddDate(0, 1, 0)
	}
	return nil
}

// UserNode represents the relationship between users and nodes
type UserNode struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	UserID uint `json:"user_id" gorm:"not null;index"`
	NodeID uint `json:"node_id" gorm:"not null;index"`

	// Relationship
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Node Node `json:"node,omitempty" gorm:"foreignKey:NodeID"`

	// Access control
	IsEnabled bool `json:"is_enabled" gorm:"not null;default:true"`
	Priority  int  `json:"priority" gorm:"not null;default:0;comment:Lower number means higher priority"`

	// Statistics
	ConnectCount int64     `json:"connect_count" gorm:"not null;default:0"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}

// TableName returns the table name for UserNode model
func (UserNode) TableName() string {
	return "user_nodes"
}