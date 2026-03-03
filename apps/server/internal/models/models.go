package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username string `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password string `gorm:"not null" json:"-"`

	Role   string `gorm:"default:'user';index;size:20" json:"role"`
	Status string `gorm:"default:'active';index;size:20" json:"status"`

	MaxTunnels int   `gorm:"default:5" json:"maxTunnels"`
	MaxTraffic int64 `gorm:"default:107374182400" json:"maxTraffic"`

	LastLoginAt *time.Time `json:"lastLoginAt"`
	LastLoginIP string     `gorm:"size:45" json:"lastLoginIp"`

	TwoFactorEnabled bool   `gorm:"default:false" json:"twoFactorEnabled"`
	TwoFactorSecret  string `gorm:"size:32" json:"-"`

	ResetToken    string     `gorm:"size:64" json:"-"`
	ResetTokenExp *time.Time `json:"-"`

	LicenseKey    string     `gorm:"size:64;index" json:"licenseKey"`
	LicenseExpire *time.Time `json:"licenseExpire"`
}

type Tunnel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint `gorm:"not null;index" json:"userId"`

	Name    string `gorm:"not null" json:"name"`
	Remark  string `gorm:"type:text" json:"remark"`
	Enabled bool   `gorm:"default:true" json:"enabled"`

	InboundProtocol string `gorm:"not null" json:"inboundProtocol"`
	InboundPort     int    `gorm:"not null" json:"inboundPort"`
	InboundListen   string `gorm:"default:'0.0.0.0'" json:"inboundListen"`
	InboundAuth     bool   `gorm:"default:false" json:"inboundAuth"`
	InboundUsername string `gorm:"type:varchar(255)" json:"inboundUsername"`
	InboundPassword string `gorm:"type:varchar(255)" json:"inboundPassword"`

	UDPEnabled bool `gorm:"default:true" json:"udpEnabled"`

	Outbounds []Outbound `gorm:"foreignKey:TunnelID" json:"outbounds"`

	RoutingStrategy string `gorm:"default:'fallback'" json:"routingStrategy"`

	UploadBytes   int64 `gorm:"default:0" json:"uploadBytes"`
	DownloadBytes int64 `gorm:"default:0" json:"downloadBytes"`
	Connections   int   `gorm:"default:0" json:"connections"`

	TrafficLimit         int64  `gorm:"default:0" json:"trafficLimit"`
	TrafficLimitUpload   int64  `gorm:"default:0" json:"trafficLimitUpload"`
	TrafficLimitDownload int64  `gorm:"default:0" json:"trafficLimitDownload"`
	TrafficResetCycle    string `gorm:"default:'monthly'" json:"trafficResetCycle"`

	SpeedLimit         int64 `gorm:"default:0" json:"speedLimit"`
	SpeedLimitUpload   int64 `gorm:"default:0" json:"speedLimitUpload"`
	SpeedLimitDownload int64 `gorm:"default:0" json:"speedLimitDownload"`

	ExpireTime *time.Time `json:"expireTime"`
	ExpireAt   *time.Time `json:"expireAt"`

	ACLEnabled bool   `gorm:"default:false" json:"aclEnabled"`
	ACLMode    string `gorm:"default:'blacklist'" json:"aclMode"` // "blacklist" or "whitelist"

	AllowDomains string `gorm:"type:text" json:"allowDomains"` // JSON array
	AllowIPs     string `gorm:"type:text" json:"allowIps"`     // JSON array
	DenyDomains  string `gorm:"type:text" json:"denyDomains"`  // JSON array
	DenyIPs      string `gorm:"type:text" json:"denyIps"`      // JSON array
}

type Outbound struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TunnelID uint   `gorm:"not null;index" json:"tunnelId"`
	Name     string `gorm:"not null" json:"name"`

	Protocol string `gorm:"not null" json:"protocol"`
	Address  string `gorm:"not null" json:"address"`
	Port     int    `gorm:"not null" json:"port"`
	Config   string `gorm:"type:text" json:"config"`

	Weight int `gorm:"default:1" json:"weight"`

	HealthCheckEnabled  bool   `gorm:"default:false" json:"healthCheckEnabled"`
	HealthCheckURL      string `json:"healthCheckUrl"`
	HealthCheckInterval int    `gorm:"default:30" json:"healthCheckInterval"`

	IsHealthy bool `gorm:"default:true" json:"isHealthy"`
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`

	UserID   uint   `gorm:"index" json:"userId"`
	Username string `gorm:"size:255" json:"username"`

	Action     string `gorm:"size:100;index" json:"action"`
	Resource   string `gorm:"size:255" json:"resource"`
	ResourceID uint   `json:"resourceId"`

	Detail string `gorm:"type:text" json:"detail"`

	IPAddress string `gorm:"size:45" json:"ipAddress"`
	UserAgent string `gorm:"type:text" json:"userAgent"`

	Success bool   `gorm:"default:true" json:"success"`
	Error   string `gorm:"type:text" json:"error"`
}

type LicenseCache struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`

	UserID      uint       `gorm:"uniqueIndex;not null" json:"userId"`
	LicenseKey  string     `gorm:"size:64" json:"licenseKey"`
	InstanceID  string     `gorm:"size:64" json:"instanceId"`
	IsValid     bool       `gorm:"default:false" json:"isValid"`
	Type        string     `gorm:"size:20" json:"type"`
	Plan        string     `gorm:"size:20" json:"plan"`
	MaxTunnels  int        `json:"maxTunnels"`
	MaxUsers    int        `json:"maxUsers"`
	MaxTraffic  int64      `json:"maxTraffic"`
	Features    string     `gorm:"type:text" json:"features"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	LastChecked *time.Time `json:"lastChecked"`
}

type SystemSettings struct {
	ID uint `gorm:"primaryKey" json:"id"`

	RegistrationEnabled bool   `gorm:"default:false" json:"registrationEnabled"`
	InviteOnly          bool   `gorm:"default:false" json:"inviteOnly"`
	IPWhitelist         string `gorm:"type:text" json:"ipWhitelist"`
	IPWhitelistEnabled  bool   `gorm:"default:false" json:"ipWhitelistEnabled"`

	SMTPHost     string `gorm:"size:255" json:"smtpHost"`
	SMTPPort     int    `json:"smtpPort"`
	SMTPUsername string `gorm:"size:255" json:"smtpUsername"`
	SMTPPassword string `gorm:"size:255" json:"smtpPassword"`
	SMTPFrom     string `gorm:"size:255" json:"smtpFrom"`
	SMTPTLS      bool   `gorm:"default:true" json:"smtpTLS"`

	TrafficAlertPercent int `gorm:"default:80" json:"trafficAlertPercent"`
	LicenseAlertDays    int `gorm:"default:7" json:"licenseAlertDays"`

	TelegramEnabled bool   `gorm:"default:false" json:"telegramEnabled"`
	TelegramToken   string `gorm:"size:255" json:"telegramToken"`
	TelegramChatID  int64  `gorm:"default:0" json:"telegramChatId"`
}

type InviteCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Code      string     `gorm:"uniqueIndex;size:32" json:"code"`
	MaxUses   int        `gorm:"default:1" json:"maxUses"`
	UsedCount int        `gorm:"default:0" json:"usedCount"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedBy uint       `json:"createdBy"`
}

type LicenseKey struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	LicenseKey string     `gorm:"uniqueIndex;size:64" json:"licenseKey"`
	Type       string     `gorm:"size:32" json:"type"`
	Plan       string     `gorm:"size:32" json:"plan"`
	MaxTunnels int        `json:"maxTunnels"`
	MaxUsers   int        `json:"maxUsers"`
	MaxTraffic int64      `json:"maxTraffic"`
	Features   string     `gorm:"size:512" json:"features"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Status     string     `gorm:"size:20;default:'unused'" json:"status"`
	UsedBy     *uint      `json:"usedBy"`
	UsedAt     *time.Time `json:"usedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}
