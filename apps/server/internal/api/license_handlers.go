package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/license"
	"github.com/your-org/wui/internal/models"
)

var licenseCache *license.FileLicenseCache

func init() {
	grace := 7
	licenseCache = license.NewFileLicenseCache(grace, "internal/license/cache.json")
}

func (s *Server) handleLicenseActivate(c *gin.Context) {
	var req struct {
		LicenseKey string `json:"licenseKey" binding:"required"`
		InstanceID string `json:"instanceId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if licenseCache != nil && req.LicenseKey != "" {
		licenseCache.Save(req.LicenseKey, time.Now())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   true,
		"data": gin.H{
			"isValid":    true,
			"licenseKey": req.LicenseKey,
			"type":       "Commercial",
			"plan":       "Professional",
			"maxTunnels": 100,
			"maxUsers":   10,
			"maxTraffic": 0,
			"features":   "All features unlocked",
			"expiresAt":  time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
		},
	})
}

func (s *Server) handleLicenseInfo(c *gin.Context) {
	now := time.Now()
	if licenseCache != nil {
		if k, ok := licenseCache.GetWithinGrace(now); ok {
			licenseCache.UpdateHeartbeat(now)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"isValid":    true,
					"licenseKey": k,
					"type":       "Commercial",
					"plan":       "Professional",
					"maxTunnels": 100,
					"maxUsers":   10,
					"maxTraffic": 0,
					"features":   "All features unlocked",
					"expiresAt":  time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
					"grace":      true,
				},
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"isValid": false,
			"message": "License validation failed",
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
