package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/license"
	"github.com/your-org/wui/internal/models"
)

var licenseCache *license.FileLicenseCache

type licenseAttempt struct {
	Count       int
	FirstTry    time.Time
	BannedUntil time.Time
}

var (
	licenseAttemptMap = make(map[string]*licenseAttempt)
	licenseAttemptMu  sync.RWMutex
)

const (
	maxLicenseAttempts   = 3
	licenseBanDuration   = time.Hour
	licenseAttemptWindow = 5 * time.Minute
)

func init() {
	grace := 7
	licenseCache = license.NewFileLicenseCache(grace, "internal/license/cache.json")

	go func() {
		for range time.Tick(5 * time.Minute) {
			cleanupLicenseAttempts()
		}
	}()
}

func cleanupLicenseAttempts() {
	licenseAttemptMu.Lock()
	defer licenseAttemptMu.Unlock()

	now := time.Now()
	for ip, attempt := range licenseAttemptMap {
		if now.After(attempt.BannedUntil) && now.Sub(attempt.FirstTry) > licenseAttemptWindow {
			delete(licenseAttemptMap, ip)
		}
	}
}

func checkLicenseRateLimit(ip string) (allowed bool, remaining int, banned bool, banRemaining time.Duration) {
	licenseAttemptMu.Lock()
	defer licenseAttemptMu.Unlock()

	now := time.Now()
	attempt, exists := licenseAttemptMap[ip]

	if !exists {
		licenseAttemptMap[ip] = &licenseAttempt{
			Count:    0,
			FirstTry: now,
		}
		return true, maxLicenseAttempts, false, 0
	}

	if now.Before(attempt.BannedUntil) {
		return false, 0, true, attempt.BannedUntil.Sub(now)
	}

	if now.Sub(attempt.FirstTry) > licenseAttemptWindow {
		attempt.Count = 0
		attempt.FirstTry = now
	}

	if attempt.Count >= maxLicenseAttempts {
		attempt.BannedUntil = now.Add(licenseBanDuration)
		return false, 0, true, licenseBanDuration
	}

	return true, maxLicenseAttempts - attempt.Count - 1, false, 0
}

func recordLicenseAttempt(ip string, success bool) {
	licenseAttemptMu.Lock()
	defer licenseAttemptMu.Unlock()

	attempt, exists := licenseAttemptMap[ip]
	if !exists {
		return
	}

	if success {
		delete(licenseAttemptMap, ip)
		return
	}

	attempt.Count++
}

func (s *Server) handleLicenseActivate(c *gin.Context) {
	ip := c.ClientIP()

	allowed, remaining, banned, banRemaining := checkLicenseRateLimit(ip)
	if banned {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success":    false,
			"valid":      false,
			"error":      "Too many failed attempts",
			"code":       "LICENSE_IP_BANNED",
			"retryAfter": int(banRemaining.Seconds()),
		})
		return
	}

	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success":   false,
			"valid":     false,
			"error":     "Rate limit exceeded",
			"code":      "LICENSE_RATE_LIMIT",
			"remaining": remaining,
		})
		return
	}

	var req struct {
		LicenseKey string `json:"licenseKey" binding:"required"`
		InstanceID string `json:"instanceId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userId")

	var lk models.LicenseKey
	if err := s.db.Where("license_key = ?", req.LicenseKey).First(&lk).Error; err != nil {
		recordLicenseAttempt(ip, false)
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"valid":     false,
			"error":     "Invalid license key",
			"code":      "LICENSE_INVALID",
			"remaining": remaining,
		})
		return
	}

	if lk.Status == "revoked" {
		recordLicenseAttempt(ip, false)
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"valid":     false,
			"error":     "License key has been revoked",
			"code":      "LICENSE_REVOKED",
			"remaining": remaining,
		})
		return
	}

	if lk.ExpiresAt != nil && time.Now().After(*lk.ExpiresAt) {
		recordLicenseAttempt(ip, false)
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"valid":     false,
			"error":     "License key has expired",
			"code":      "LICENSE_EXPIRED",
			"remaining": remaining,
		})
		return
	}

	if lk.Status == "used" && lk.UsedBy != nil && *lk.UsedBy != userID {
		recordLicenseAttempt(ip, false)
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"valid":     false,
			"error":     "License key is already used by another user",
			"code":      "LICENSE_USED",
			"remaining": remaining,
		})
		return
	}

	recordLicenseAttempt(ip, true)

	now := time.Now()
	updates := map[string]interface{}{
		"status":  "used",
		"used_by": userID,
		"used_at": now,
	}
	s.db.Model(&lk).Updates(updates)

	var user models.User
	s.db.First(&user, userID)

	if lk.MaxTunnels > 0 {
		s.db.Model(&user).Update("max_tunnels", lk.MaxTunnels)
	}
	if lk.MaxTraffic > 0 {
		s.db.Model(&user).Update("max_traffic", lk.MaxTraffic)
	}
	s.db.Model(&user).Update("license_key", lk.LicenseKey)
	if lk.ExpiresAt != nil {
		s.db.Model(&user).Update("license_expire", lk.ExpiresAt)
	}

	if licenseCache != nil {
		licenseCache.Save(req.LicenseKey, now)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   true,
		"data": gin.H{
			"isValid":    true,
			"licenseKey": lk.LicenseKey,
			"type":       lk.Type,
			"plan":       lk.Plan,
			"maxTunnels": lk.MaxTunnels,
			"maxUsers":   lk.MaxUsers,
			"maxTraffic": lk.MaxTraffic,
			"features":   lk.Features,
			"expiresAt":  lk.ExpiresAt,
		},
	})
}

func (s *Server) handleLicenseDeactivate(c *gin.Context) {
	userID := c.GetUint("userId")

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	if user.LicenseKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "No license activated",
		})
		return
	}

	var lk models.LicenseKey
	if err := s.db.Where("license_key = ?", user.LicenseKey).First(&lk).Error; err == nil {
		if lk.UsedBy != nil && *lk.UsedBy == userID {
			s.db.Model(&lk).Updates(map[string]interface{}{
				"status":  "unused",
				"used_by": nil,
				"used_at": nil,
			})
		}
	}

	s.db.Model(&user).Updates(map[string]interface{}{
		"license_key":    "",
		"license_expire": nil,
		"max_tunnels":    5,
		"max_traffic":    107374182400,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "License deactivated successfully",
	})
}

func (s *Server) handleLicenseInfo(c *gin.Context) {
	userID := c.GetUint("userId")

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"isValid": false,
				"message": "User not found",
			},
		})
		return
	}

	if user.LicenseKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"isValid": false,
				"message": "No license activated",
			},
		})
		return
	}

	var lk models.LicenseKey
	if err := s.db.Where("license_key = ?", user.LicenseKey).First(&lk).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"isValid":    false,
				"licenseKey": user.LicenseKey,
				"message":    "License not found in database",
			},
		})
		return
	}

	if lk.Status == "revoked" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"isValid":    false,
				"licenseKey": user.LicenseKey,
				"message":    "License has been revoked",
			},
		})
		return
	}

	if lk.ExpiresAt != nil && time.Now().After(*lk.ExpiresAt) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"isValid":    false,
				"licenseKey": user.LicenseKey,
				"message":    "License has expired",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"isValid":    true,
			"licenseKey": lk.LicenseKey,
			"type":       lk.Type,
			"plan":       lk.Plan,
			"maxTunnels": lk.MaxTunnels,
			"maxUsers":   lk.MaxUsers,
			"maxTraffic": lk.MaxTraffic,
			"features":   lk.Features,
			"expiresAt":  lk.ExpiresAt,
		},
	})
}

func (s *Server) handleAdminListLicenses(c *gin.Context) {
	var licenses []models.LicenseKey
	if err := s.db.Order("created_at desc").Find(&licenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to list licenses"))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(licenses))
}

func (s *Server) handleAdminCreateLicense(c *gin.Context) {
	var req struct {
		LicenseKey string     `json:"licenseKey" binding:"required"`
		Type       string     `json:"type" binding:"required"`
		Plan       string     `json:"plan" binding:"required"`
		MaxTunnels int        `json:"maxTunnels"`
		MaxUsers   int        `json:"maxUsers"`
		MaxTraffic int64      `json:"maxTraffic"`
		Features   string     `json:"features"`
		ExpiresAt  *time.Time `json:"expiresAt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	lk := models.LicenseKey{
		LicenseKey: req.LicenseKey,
		Type:       req.Type,
		Plan:       req.Plan,
		MaxTunnels: req.MaxTunnels,
		MaxUsers:   req.MaxUsers,
		MaxTraffic: req.MaxTraffic,
		Features:   req.Features,
		ExpiresAt:  req.ExpiresAt,
		Status:     "unused",
	}

	if err := s.db.Create(&lk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create license: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(lk))
}

func (s *Server) handleAdminDeleteLicense(c *gin.Context) {
	id := c.Param("id")

	if err := s.db.Delete(&models.LicenseKey{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to delete license"))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("License deleted"))
}

func (s *Server) handleAdminRevokeLicense(c *gin.Context) {
	id := c.Param("id")

	var lk models.LicenseKey
	if err := s.db.First(&lk, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("License not found"))
		return
	}

	updates := map[string]interface{}{
		"status":  "revoked",
		"used_by": nil,
		"used_at": nil,
	}

	if err := s.db.Model(&lk).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to revoke license"))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("License revoked"))
}

func (s *Server) handleAdminUpdateLicense(c *gin.Context) {
	id := c.Param("id")

	var lk models.LicenseKey
	if err := s.db.First(&lk, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("License not found"))
		return
	}

	var req struct {
		Type       string     `json:"type"`
		Plan       string     `json:"plan"`
		MaxTunnels int        `json:"maxTunnels"`
		MaxUsers   int        `json:"maxUsers"`
		MaxTraffic int64      `json:"maxTraffic"`
		Features   string     `json:"features"`
		ExpiresAt  *time.Time `json:"expiresAt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	updates := map[string]interface{}{
		"type":        req.Type,
		"plan":        req.Plan,
		"max_tunnels": req.MaxTunnels,
		"max_users":   req.MaxUsers,
		"max_traffic": req.MaxTraffic,
		"features":    req.Features,
		"expires_at":  req.ExpiresAt,
	}

	if err := s.db.Model(&lk).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to update license"))
		return
	}

	s.db.First(&lk, id)
	c.JSON(http.StatusOK, SuccessResponse(lk))
}

func GenerateLicenseKey() string {
	b := make([]byte, 6)
	rand.Read(b)
	return "WUI-" + hex.EncodeToString(b)
}
