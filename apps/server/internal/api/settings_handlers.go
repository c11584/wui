package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/models"
	"github.com/your-org/wui/internal/security"
)

func (s *Server) handleGetSettings(c *gin.Context) {
	var settings models.SystemSettings
	result := s.db.First(&settings)
	if result.Error != nil {
		settings = models.SystemSettings{
			RegistrationEnabled: false,
			InviteOnly:          false,
			IPWhitelistEnabled:  false,
			TrafficAlertPercent: 80,
			LicenseAlertDays:    7,
		}
	}

	c.JSON(http.StatusOK, SuccessResponse(settings))
}

func (s *Server) handleUpdateSettings(c *gin.Context) {
	var req models.SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	var settings models.SystemSettings
	result := s.db.First(&settings)
	if result.Error != nil {
		settings = req
		s.db.Create(&settings)
	} else {
		settings.RegistrationEnabled = req.RegistrationEnabled
		settings.InviteOnly = req.InviteOnly
		settings.IPWhitelist = req.IPWhitelist
		settings.IPWhitelistEnabled = req.IPWhitelistEnabled
		settings.SMTPHost = req.SMTPHost
		settings.SMTPPort = req.SMTPPort
		settings.SMTPUsername = req.SMTPUsername
		settings.SMTPPassword = req.SMTPPassword
		settings.SMTPFrom = req.SMTPFrom
		settings.TrafficAlertPercent = req.TrafficAlertPercent
		settings.LicenseAlertDays = req.LicenseAlertDays
		s.db.Save(&settings)
	}

	var ips []string
	if settings.IPWhitelist != "" {
		json.Unmarshal([]byte(settings.IPWhitelist), &ips)
	}
	security.GlobalIPWhitelist.Enable(settings.IPWhitelistEnabled)
	security.GlobalIPWhitelist.SetWhitelist(ips)

	c.JSON(http.StatusOK, SuccessResponse(settings))
}

func (s *Server) handleListInviteCodes(c *gin.Context) {
	var codes []models.InviteCode
	s.db.Order("created_at desc").Find(&codes)

	c.JSON(http.StatusOK, SuccessResponse(codes))
}

func (s *Server) handleCreateInviteCode(c *gin.Context) {
	userID, _ := c.Get("userId")

	var req struct {
		MaxUses   int        `json:"maxUses"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	c.ShouldBindJSON(&req)

	if req.MaxUses <= 0 {
		req.MaxUses = 1
	}

	bytes := make([]byte, 16)
	rand.Read(bytes)
	code := hex.EncodeToString(bytes)

	invite := models.InviteCode{
		Code:      code,
		MaxUses:   req.MaxUses,
		ExpiresAt: req.ExpiresAt,
		CreatedBy: userID.(uint),
	}

	if err := s.db.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create invite code"))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(invite))
}

func (s *Server) handleDeleteInviteCode(c *gin.Context) {
	id := c.Param("id")
	if err := s.db.Delete(&models.InviteCode{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to delete invite code"))
		return
	}
	c.JSON(http.StatusOK, SuccessMessage("Invite code deleted"))
}

func (s *Server) handleBackup(c *gin.Context) {
	var users []models.User
	var tunnels []models.Tunnel
	var settings models.SystemSettings

	s.db.Find(&users)
	s.db.Preload("Outbounds").Find(&tunnels)
	s.db.First(&settings)

	backup := gin.H{
		"timestamp": time.Now().Unix(),
		"users":     users,
		"tunnels":   tunnels,
		"settings":  settings,
	}

	c.JSON(http.StatusOK, backup)
}

func (s *Server) handleRestore(c *gin.Context) {
	var backup struct {
		Users    []models.User         `json:"users"`
		Tunnels  []models.Tunnel       `json:"tunnels"`
		Settings models.SystemSettings `json:"settings"`
	}

	if err := c.ShouldBindJSON(&backup); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid backup data"))
		return
	}

	tx := s.db.Begin()

	for _, user := range backup.Users {
		tx.FirstOrCreate(&user, models.User{Username: user.Username})
	}

	for _, tunnel := range backup.Tunnels {
		tunnel.ID = 0
		for i := range tunnel.Outbounds {
			tunnel.Outbounds[i].ID = 0
			tunnel.Outbounds[i].TunnelID = 0
		}
		tx.Create(&tunnel)
	}

	tx.Commit()

	c.JSON(http.StatusOK, SuccessMessage("Restore completed"))
}
