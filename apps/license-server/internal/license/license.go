package license

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/your-org/wui/apps/license-server/internal/models"
	"gorm.io/gorm"
)

const (
	charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 去除易混淆字符 0O1IL
)

// Generator handles license key generation
type Generator struct {
	db *gorm.DB
}

// NewGenerator creates a new license generator
func NewGenerator(db *gorm.DB) *Generator {
	return &Generator{db: db}
}

// GenerateKey generates a new license key
func (g *Generator) GenerateKey() (string, error) {
	for i := 0; i < 100; i++ { // 最多尝试 100 次
		key := generateRandomKey()
		hash := HashLicenseKey(key)

		// 检查是否已存在
		var count int64
		g.db.Model(&models.License{}).Where("license_hash = ?", hash).Count(&count)
		if count == 0 {
			return key, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique key after 100 attempts")
}

// generateRandomKey generates a random license key
func generateRandomKey() string {
	bytes := make([]byte, 12) // 3 groups of 4 chars = 12 bytes
	rand.Read(bytes)

	result := "WUI"
	for i := 0; i < 12; i++ {
		if i > 0 && i%4 == 0 {
			result += "-"
		}
		result += string(charset[int(bytes[i])%len(charset)])
	}

	return result
}

// HashLicenseKey creates a SHA256 hash of the license key
func HashLicenseKey(key string) string {
	key = strings.ToUpper(strings.TrimSpace(key))
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// CreateLicense creates a new license with the given parameters
func (g *Generator) CreateLicense(licenseType, plan string, maxTunnels, maxUsers int, maxTraffic int64, expiresAt *time.Time, customerName, customerEmail, orderID, remark string) (*models.License, error) {
	key, err := g.GenerateKey()
	if err != nil {
		return nil, err
	}

	license := &models.License{
		LicenseKey:    key,
		LicenseHash:   HashLicenseKey(key),
		Type:          licenseType,
		Plan:          plan,
		MaxTunnels:    maxTunnels,
		MaxUsers:      maxUsers,
		MaxTraffic:    maxTraffic,
		ExpiresAt:     expiresAt,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		OrderID:       orderID,
		Remark:        remark,
	}

	if err := g.db.Create(license).Error; err != nil {
		return nil, err
	}

	return license, nil
}

// Validator handles license validation
type Validator struct {
	db *gorm.DB
}

// NewValidator creates a new license validator
func NewValidator(db *gorm.DB) *Validator {
	return &Validator{db: db}
}

// Validate validates a license key
func (v *Validator) Validate(req *models.ValidateRequest) (*models.ValidateResponse, error) {
	hash := HashLicenseKey(req.LicenseKey)

	var license models.License
	if err := v.db.Where("license_hash = ?", hash).First(&license).Error; err != nil {
		return &models.ValidateResponse{
			Valid:   false,
			Message: "Invalid license key",
		}, nil
	}

	// 检查状态
	if license.Status == "suspended" {
		return &models.ValidateResponse{
			Valid:   false,
			Message: "License is suspended",
		}, nil
	}

	if license.Status == "expired" {
		return &models.ValidateResponse{
			Valid:   false,
			Message: "License has expired",
		}, nil
	}

	// 检查有效期
	if !license.Lifetime && license.ExpiresAt != nil && license.ExpiresAt.Before(time.Now()) {
		// 更新状态为过期
		v.db.Model(&license).Update("status", "expired")
		return &models.ValidateResponse{
			Valid:   false,
			Message: "License has expired",
		}, nil
	}

	// 检查实例绑定
	if license.InstanceID != "" && license.InstanceID != req.InstanceID {
		return &models.ValidateResponse{
			Valid:   false,
			Message: "License is bound to another instance",
		}, nil
	}

	// 更新最后检查时间
	now := time.Now()
	v.db.Model(&license).Update("last_check_at", now)

	expiresAtStr := ""
	if license.ExpiresAt != nil {
		expiresAtStr = license.ExpiresAt.Format(time.RFC3339)
	}

	return &models.ValidateResponse{
		Valid:      true,
		Message:    "License is valid",
		Type:       license.Type,
		Plan:       license.Plan,
		MaxTunnels: license.MaxTunnels,
		MaxUsers:   license.MaxUsers,
		MaxTraffic: license.MaxTraffic,
		Features:   license.Features,
		ExpiresAt:  expiresAtStr,
	}, nil
}

// Activate activates a license for a specific instance
func (v *Validator) Activate(req *models.ActivateRequest) (*models.ValidateResponse, error) {
	hash := HashLicenseKey(req.LicenseKey)

	var license models.License
	if err := v.db.Where("license_hash = ?", hash).First(&license).Error; err != nil {
		return &models.ValidateResponse{
			Valid:   false,
			Message: "Invalid license key",
		}, nil
	}

	// 检查是否已激活
	if license.InstanceID != "" && license.InstanceID != req.InstanceID {
		return &models.ValidateResponse{
			Valid:   false,
			Message: "License is already activated on another instance",
		}, nil
	}

	// 激活
	now := time.Now()
	updates := map[string]interface{}{
		"instance_id":     req.InstanceID,
		"bind_machine_id": req.MachineID,
		"bind_domain":     req.Domain,
		"bind_ip":         req.IPAddress,
		"status":          "active",
		"activated_at":    now,
		"last_check_at":   now,
	}

	if err := v.db.Model(&license).Updates(updates).Error; err != nil {
		return nil, err
	}

	expiresAtStr := ""
	if license.ExpiresAt != nil {
		expiresAtStr = license.ExpiresAt.Format(time.RFC3339)
	}

	return &models.ValidateResponse{
		Valid:      true,
		Message:    "License activated successfully",
		Type:       license.Type,
		Plan:       license.Plan,
		MaxTunnels: license.MaxTunnels,
		MaxUsers:   license.MaxUsers,
		MaxTraffic: license.MaxTraffic,
		Features:   license.Features,
		ExpiresAt:  expiresAtStr,
	}, nil
}

// RecordHeartbeat records a heartbeat from a license instance
func (v *Validator) RecordHeartbeat(req *models.HeartbeatRequest) error {
	heartbeat := &models.Heartbeat{
		LicenseKey:  req.LicenseKey,
		InstanceID:  req.InstanceID,
		Status:      "ok",
		Version:     req.Version,
		TunnelCount: req.TunnelCount,
		UserCount:   req.UserCount,
		CpuUsage:    req.CpuUsage,
		MemUsage:    req.MemUsage,
		DiskUsage:   req.DiskUsage,
		IPAddress:   req.IPAddress,
		Domain:      req.Domain,
	}

	return v.db.Create(heartbeat).Error
}
