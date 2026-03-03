package models

import (
	"time"
)

// License represents a license key
type License struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 密钥信息
	LicenseKey  string `gorm:"uniqueIndex;not null;size:19" json:"licenseKey"` // WUI-XXXX-XXXX-XXXX
	LicenseHash string `gorm:"uniqueIndex;size:64" json:"-"`                   // SHA256 哈希

	// 类型
	Type string `gorm:"not null;index;size:20" json:"type"` // trial, personal, team, enterprise
	Plan string `gorm:"not null;size:20" json:"plan"`       // basic, pro, enterprise

	// 功能限制
	MaxTunnels int    `gorm:"default:5" json:"maxTunnels"`
	MaxUsers   int    `gorm:"default:1" json:"maxUsers"`
	MaxTraffic int64  `gorm:"default:107374182400" json:"maxTraffic"` // bytes
	Features   string `gorm:"type:text" json:"features"`              // JSON array

	// 有效期
	ExpiresAt *time.Time `json:"expiresAt"`
	Lifetime  bool       `gorm:"default:false" json:"lifetime"`

	// 绑定信息
	InstanceID    string `gorm:"index;size:64" json:"instanceId"`
	BindIP        string `gorm:"size:45" json:"bindIp"`
	BindDomain    string `gorm:"size:255" json:"bindDomain"`
	BindMachineID string `gorm:"size:64" json:"bindMachineId"`

	// 状态
	Status      string     `gorm:"default:'inactive';index;size:20" json:"status"` // inactive, active, suspended, expired
	ActivatedAt *time.Time `json:"activatedAt"`
	LastCheckAt *time.Time `json:"lastCheckAt"`

	// 客户信息
	CustomerName  string `gorm:"size:255" json:"customerName"`
	CustomerEmail string `gorm:"size:255" json:"customerEmail"`
	OrderID       string `gorm:"size:100" json:"orderId"`

	// 备注
	Remark string `gorm:"type:text" json:"remark"`
}

// Heartbeat represents a license heartbeat record
type Heartbeat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`

	LicenseKey string `gorm:"index;size:19" json:"licenseKey"`
	InstanceID string `gorm:"index;size:64" json:"instanceId"`

	// 状态
	Status      string `gorm:"size:20" json:"status"` // ok, warning, error
	Version     string `gorm:"size:20" json:"version"`
	TunnelCount int    `json:"tunnelCount"`
	UserCount   int    `json:"userCount"`

	// 资源使用
	CpuUsage  float64 `json:"cpuUsage"`
	MemUsage  float64 `json:"memUsage"`
	DiskUsage float64 `json:"diskUsage"`

	// 网络
	IPAddress string `gorm:"size:45" json:"ipAddress"`
	Domain    string `gorm:"size:255" json:"domain"`
}

// ValidateRequest represents a license validation request
type ValidateRequest struct {
	LicenseKey string `json:"licenseKey" binding:"required"`
	InstanceID string `json:"instanceId" binding:"required"`
	MachineID  string `json:"machineId"`
	Domain     string `json:"domain"`
	IPAddress  string `json:"ipAddress"`
}

// ValidateResponse represents a license validation response
type ValidateResponse struct {
	Valid      bool   `json:"valid"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	Plan       string `json:"plan"`
	MaxTunnels int    `json:"maxTunnels"`
	MaxUsers   int    `json:"maxUsers"`
	MaxTraffic int64  `json:"maxTraffic"`
	Features   string `json:"features"`
	ExpiresAt  string `json:"expiresAt"`
}

// ActivateRequest represents a license activation request
type ActivateRequest struct {
	LicenseKey string `json:"licenseKey" binding:"required"`
	InstanceID string `json:"instanceId" binding:"required"`
	MachineID  string `json:"machineId"`
	Domain     string `json:"domain"`
	IPAddress  string `json:"ipAddress"`
}

// HeartbeatRequest represents a heartbeat request
type HeartbeatRequest struct {
	LicenseKey  string  `json:"licenseKey" binding:"required"`
	InstanceID  string  `json:"instanceId" binding:"required"`
	Version     string  `json:"version"`
	TunnelCount int     `json:"tunnelCount"`
	UserCount   int     `json:"userCount"`
	CpuUsage    float64 `json:"cpuUsage"`
	MemUsage    float64 `json:"memUsage"`
	DiskUsage   float64 `json:"diskUsage"`
	IPAddress   string  `json:"ipAddress"`
	Domain      string  `json:"domain"`
}
