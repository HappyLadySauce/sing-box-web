package models

import (
	"time"

	"gorm.io/gorm"
)

// NodeStatus represents node status
type NodeStatus string

const (
	NodeStatusOnline     NodeStatus = "online"
	NodeStatusOffline    NodeStatus = "offline"
	NodeStatusMaintenance NodeStatus = "maintenance"
	NodeStatusDisabled   NodeStatus = "disabled"
)

// NodeType represents node type
type NodeType string

const (
	NodeTypeVMess      NodeType = "vmess"
	NodeTypeVLESS      NodeType = "vless"
	NodeTypeTrojan     NodeType = "trojan"
	NodeTypeShadowsocks NodeType = "shadowsocks"
	NodeTypeHysteria   NodeType = "hysteria"
	NodeTypeHysteria2  NodeType = "hysteria2"
	NodeTypeTUIC       NodeType = "tuic"
)

// Node represents a sing-box server node
type Node struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Basic information
	Name        string     `json:"name" gorm:"not null;size:128"`
	Description string     `json:"description" gorm:"type:text"`
	Type        NodeType   `json:"type" gorm:"not null;size:20"`
	Status      NodeStatus `json:"status" gorm:"not null;default:'offline';size:20"`

	// Connection information
	Host string `json:"host" gorm:"not null;size:255"`
	Port int    `json:"port" gorm:"not null"`

	// Authentication and encryption
	UUID       string `json:"uuid,omitempty" gorm:"size:36;comment:For VMess/VLESS"`
	Password   string `json:"password,omitempty" gorm:"size:255;comment:For Trojan/Shadowsocks"`
	Method     string `json:"method,omitempty" gorm:"size:32;comment:Encryption method"`
	Protocol   string `json:"protocol,omitempty" gorm:"size:32;comment:Transport protocol"`
	
	// Transport configuration
	Network     string `json:"network,omitempty" gorm:"size:16;default:'tcp';comment:tcp/udp/ws/grpc"`
	Path        string `json:"path,omitempty" gorm:"size:255;comment:WebSocket path or gRPC service name"`
	Host_header string `json:"host_header,omitempty" gorm:"size:255;column:host_header;comment:Host header for disguise"`
	
	// TLS configuration
	TLS         bool   `json:"tls" gorm:"not null;default:false"`
	ServerName  string `json:"server_name,omitempty" gorm:"size:255;comment:TLS server name"`
	Fingerprint string `json:"fingerprint,omitempty" gorm:"size:64;comment:TLS fingerprint"`
	ALPN        string `json:"alpn,omitempty" gorm:"size:255;comment:ALPN protocols"`
	AllowInsecure bool `json:"allow_insecure" gorm:"not null;default:false"`

	// Node configuration
	MaxUsers    int   `json:"max_users" gorm:"not null;default:0;comment:0 means unlimited"`
	SpeedLimit  int64 `json:"speed_limit" gorm:"not null;default:0;comment:Speed limit per user in bytes/sec"`
	TrafficRate float64 `json:"traffic_rate" gorm:"not null;default:1.0;comment:Traffic rate multiplier"`

	// Node management
	Region      string `json:"region" gorm:"size:64"`
	Country     string `json:"country" gorm:"size:64"`
	City        string `json:"city" gorm:"size:64"`
	ISP         string `json:"isp" gorm:"size:128"`
	Tags        string `json:"tags" gorm:"size:512;comment:Comma-separated tags"`
	Sort        int    `json:"sort" gorm:"not null;default:0;comment:Sort order"`
	IsEnabled   bool   `json:"is_enabled" gorm:"not null;default:true"`

	// Statistics and monitoring
	CurrentUsers   int       `json:"current_users" gorm:"not null;default:0"`
	TotalTraffic   int64     `json:"total_traffic" gorm:"not null;default:0;comment:Total traffic in bytes"`
	UploadTraffic  int64     `json:"upload_traffic" gorm:"not null;default:0"`
	DownloadTraffic int64    `json:"download_traffic" gorm:"not null;default:0"`
	LastHeartbeat  *time.Time `json:"last_heartbeat,omitempty"`
	
	// System information
	CPUUsage    float64 `json:"cpu_usage" gorm:"type:decimal(5,2);default:0"`
	MemoryUsage float64 `json:"memory_usage" gorm:"type:decimal(5,2);default:0"`
	DiskUsage   float64 `json:"disk_usage" gorm:"type:decimal(5,2);default:0"`
	Load1       float64 `json:"load1" gorm:"type:decimal(8,2);default:0"`
	Load5       float64 `json:"load5" gorm:"type:decimal(8,2);default:0"`
	Load15      float64 `json:"load15" gorm:"type:decimal(8,2);default:0"`

	// Configuration and version
	ConfigVersion  int    `json:"config_version" gorm:"not null;default:0"`
	AgentVersion   string `json:"agent_version" gorm:"size:32"`
	SingBoxVersion string `json:"sing_box_version" gorm:"size:32"`

	// Metadata
	Notes    string            `json:"notes" gorm:"type:text"`
	Metadata map[string]string `json:"metadata,omitempty" gorm:"serializer:json"`

	// Relationships
	TrafficRecords []TrafficRecord `json:"traffic_records,omitempty" gorm:"foreignKey:NodeID"`
	UserNodes      []UserNode      `json:"user_nodes,omitempty" gorm:"foreignKey:NodeID"`
	NodeLogs       []NodeLog       `json:"node_logs,omitempty" gorm:"foreignKey:NodeID"`
}

// TableName returns the table name for Node model
func (Node) TableName() string {
	return "nodes"
}

// IsOnline checks if node is online
func (n *Node) IsOnline() bool {
	if n.Status != NodeStatusOnline {
		return false
	}
	if n.LastHeartbeat == nil {
		return false
	}
	// Consider node offline if no heartbeat for 2 minutes
	return time.Since(*n.LastHeartbeat) < 2*time.Minute
}

// IsAvailable checks if node is available for users
func (n *Node) IsAvailable() bool {
	return n.IsEnabled && n.IsOnline() && n.Status != NodeStatusMaintenance
}

// CanAcceptNewUser checks if node can accept new users
func (n *Node) CanAcceptNewUser() bool {
	if !n.IsAvailable() {
		return false
	}
	if n.MaxUsers <= 0 {
		return true // Unlimited
	}
	return n.CurrentUsers < n.MaxUsers
}

// GetUsagePercentage returns the usage percentage
func (n *Node) GetUsagePercentage() float64 {
	if n.MaxUsers <= 0 {
		return 0 // Unlimited
	}
	return float64(n.CurrentUsers) / float64(n.MaxUsers) * 100
}

// UpdateHeartbeat updates the last heartbeat time
func (n *Node) UpdateHeartbeat() {
	now := time.Now()
	n.LastHeartbeat = &now
}

// UpdateSystemInfo updates system monitoring information
func (n *Node) UpdateSystemInfo(cpu, memory, disk, load1, load5, load15 float64) {
	n.CPUUsage = cpu
	n.MemoryUsage = memory
	n.DiskUsage = disk
	n.Load1 = load1
	n.Load5 = load5
	n.Load15 = load15
}

// UpdateTraffic updates traffic statistics
func (n *Node) UpdateTraffic(upload, download int64) {
	n.UploadTraffic += upload
	n.DownloadTraffic += download
	n.TotalTraffic = n.UploadTraffic + n.DownloadTraffic
}

// BeforeCreate GORM hook to set defaults before creating
func (n *Node) BeforeCreate(tx *gorm.DB) error {
	if n.UUID == "" && (n.Type == NodeTypeVMess || n.Type == NodeTypeVLESS) {
		n.UUID = generateUUID()
	}
	return nil
}

// NodeLog represents node operation logs
type NodeLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	NodeID uint `json:"node_id" gorm:"not null;index"`
	Node   Node `json:"node,omitempty" gorm:"foreignKey:NodeID"`

	Level   string `json:"level" gorm:"not null;size:10;index"`
	Type    string `json:"type" gorm:"not null;size:32;index;comment:heartbeat/traffic/system/error"`
	Message string `json:"message" gorm:"not null;type:text"`
	
	// Additional data
	Data map[string]interface{} `json:"data,omitempty" gorm:"serializer:json"`
}

// TableName returns the table name for NodeLog model
func (NodeLog) TableName() string {
	return "node_logs"
}