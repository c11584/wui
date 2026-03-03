package models

import (
	"gorm.io/gorm"
	"time"
)

type Admin struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Email     string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Role      string         `gorm:"default:'admin';size:20" json:"role"`
}

type Customer struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"not null;size:255" json:"name"`
	Email     string         `gorm:"size:255" json:"email"`
	Phone     string         `gorm:"size:50" json:"phone"`
	Company   string         `gorm:"size:255" json:"company"`
	Notes     string         `gorm:"type:text" json:"notes"`
	Status    string         `gorm:"default:'active';size:20" json:"status"`
}

type License struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Key           string         `gorm:"uniqueIndex;not null;size:64" json:"key"`
	Hash          string         `gorm:"not null;size:64" json:"-"`
	Type          string         `gorm:"not null;size:20" json:"type"`
	Plan          string         `gorm:"not null;size:20" json:"plan"`
	Status        string         `gorm:"default:'inactive';size:20" json:"status"`
	MaxTunnels    int            `gorm:"default:5" json:"maxTunnels"`
	MaxUsers      int            `gorm:"default:1" json:"maxUsers"`
	MaxTraffic    int64          `gorm:"default:107374182400" json:"maxTraffic"`
	Features      string         `gorm:"type:text" json:"features"`
	ExpiresAt     *time.Time     `json:"expiresAt"`
	ActivatedAt   *time.Time     `json:"activatedAt"`
	ActivatedBy   string         `gorm:"size:64" json:"activatedBy"`
	LastHeartbeat *time.Time     `json:"lastHeartbeat"`
	CustomerID    *uint          `json:"customerId"`
	Customer      *Customer      `json:"customer,omitempty"`
}

type HeartbeatLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
	LicenseID   uint      `gorm:"index" json:"licenseId"`
	LicenseKey  string    `gorm:"size:64" json:"licenseKey"`
	InstanceID  string    `gorm:"size:64" json:"instanceId"`
	IPAddress   string    `gorm:"size:45" json:"ipAddress"`
	Domain      string    `gorm:"size:255" json:"domain"`
	TunnelCount int       `json:"tunnelCount"`
	UserCount   int       `json:"userCount"`
	CPUUsage    float64   `json:"cpuUsage"`
	MemUsage    float64   `json:"memUsage"`
	DiskUsage   float64   `json:"diskUsage"`
}

type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time `gorm:"index" json:"createdAt"`
	AdminID    uint      `gorm:"index" json:"adminId"`
	AdminName  string    `gorm:"size:255" json:"adminName"`
	Action     string    `gorm:"size:100;index" json:"action"`
	Resource   string    `gorm:"size:255" json:"resource"`
	ResourceID uint      `json:"resourceId"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IPAddress  string    `gorm:"size:45" json:"ipAddress"`
	UserAgent  string    `gorm:"type:text" json:"userAgent"`
	Success    bool      `gorm:"default:true" json:"success"`
	Error      string    `gorm:"type:text" json:"error"`
}
