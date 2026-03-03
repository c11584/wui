package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/models"
)

func (s *Server) handleListTunnels(c *gin.Context) {
	name := c.Query("name")
	protocol := c.Query("protocol")
	enabled := c.Query("enabled")
	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	query := s.db.Preload("Outbounds")

	// Non-admin users can only see their own tunnels
	if userRole != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if protocol != "" {
		query = query.Where("inbound_protocol = ?", protocol)
	}
	if enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}

	var tunnels []models.Tunnel
	if err := query.Find(&tunnels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to fetch tunnels"))
		return
	}

	type TunnelWithStatus struct {
		models.Tunnel
		IsRunning bool `json:"isRunning"`
	}

	var result []TunnelWithStatus
	for _, t := range tunnels {
		result = append(result, TunnelWithStatus{
			Tunnel:    t,
			IsRunning: s.tunnelMgr.IsRunning(t.ID),
		})
	}

	c.JSON(http.StatusOK, SuccessResponse(result))
}

func (s *Server) handleGetTunnel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid tunnel ID"))
		return
	}

	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	var tunnel models.Tunnel
	if err := s.db.Preload("Outbounds").First(&tunnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Tunnel not found"))
		return
	}

	if userRole != "admin" && tunnel.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse("Access denied"))
		return
	}

	type TunnelWithStatus struct {
		models.Tunnel
		IsRunning bool `json:"isRunning"`
	}

	c.JSON(http.StatusOK, SuccessResponse(TunnelWithStatus{
		Tunnel:    tunnel,
		IsRunning: s.tunnelMgr.IsRunning(tunnel.ID),
	}))
}

func (s *Server) handleCreateTunnel(c *gin.Context) {
	userID := c.GetUint("userId")

	var tunnel models.Tunnel
	if err := c.ShouldBindJSON(&tunnel); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	if err := s.tunnelMgr.ValidateConfig(&tunnel); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	var count int64
	s.db.Model(&models.Tunnel{}).Where("inbound_port = ?", tunnel.InboundPort).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse("Port already in use"))
		return
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("User not found"))
		return
	}

	s.db.Model(&models.Tunnel{}).Where("user_id = ?", userID).Count(&count)
	if user.MaxTunnels > 0 && int(count) >= user.MaxTunnels {
		c.JSON(http.StatusForbidden, ErrorResponse(fmt.Sprintf("Maximum tunnel limit reached (%d)", user.MaxTunnels)))
		return
	}

	tunnel.UserID = userID

	if err := s.db.Create(&tunnel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create tunnel"))
		return
	}

	s.db.Preload("Outbounds").First(&tunnel, tunnel.ID)

	s.userHandlers.logAudit(c, userID, user.Username, "create_tunnel", "tunnels", int(tunnel.ID), fmt.Sprintf("Created tunnel: %s", tunnel.Name), true)

	c.JSON(http.StatusCreated, SuccessResponse(tunnel))
}

func (s *Server) handleUpdateTunnel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid tunnel ID"))
		return
	}

	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	var tunnel models.Tunnel
	if err := s.db.First(&tunnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Tunnel not found"))
		return
	}

	if userRole != "admin" && tunnel.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse("Access denied"))
		return
	}

	var req models.Tunnel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	if err := s.tunnelMgr.ValidateConfig(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if req.InboundPort != tunnel.InboundPort {
		var count int64
		s.db.Model(&models.Tunnel{}).Where("inbound_port = ? AND id != ?", req.InboundPort, tunnel.ID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse("Port already in use"))
			return
		}
	}

	tunnel.Name = req.Name
	tunnel.Remark = req.Remark
	tunnel.Enabled = req.Enabled
	tunnel.InboundProtocol = req.InboundProtocol
	tunnel.InboundPort = req.InboundPort
	tunnel.InboundListen = req.InboundListen
	tunnel.InboundAuth = req.InboundAuth
	tunnel.InboundUsername = req.InboundUsername
	tunnel.InboundPassword = req.InboundPassword
	tunnel.UDPEnabled = req.UDPEnabled
	tunnel.RoutingStrategy = req.RoutingStrategy
	tunnel.TrafficLimitUpload = req.TrafficLimitUpload
	tunnel.TrafficLimitDownload = req.TrafficLimitDownload
	tunnel.TrafficResetCycle = req.TrafficResetCycle
	tunnel.SpeedLimitUpload = req.SpeedLimitUpload
	tunnel.SpeedLimitDownload = req.SpeedLimitDownload
	tunnel.ExpireTime = req.ExpireTime
	tunnel.ACLEnabled = req.ACLEnabled
	tunnel.ACLMode = req.ACLMode
	tunnel.AllowDomains = req.AllowDomains
	tunnel.AllowIPs = req.AllowIPs
	tunnel.DenyDomains = req.DenyDomains
	tunnel.DenyIPs = req.DenyIPs

	s.db.Where("tunnel_id = ?", tunnel.ID).Delete(&models.Outbound{})
	tunnel.Outbounds = req.Outbounds

	if err := s.db.Save(&tunnel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to update tunnel"))
		return
	}

	s.db.Preload("Outbounds").First(&tunnel, tunnel.ID)

	c.JSON(http.StatusOK, SuccessResponse(tunnel))
}

func (s *Server) handleDeleteTunnel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid tunnel ID"))
		return
	}

	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	var tunnel models.Tunnel
	if err := s.db.First(&tunnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Tunnel not found"))
		return
	}

	if userRole != "admin" && tunnel.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse("Access denied"))
		return
	}

	s.tunnelMgr.StopTunnel(tunnel.ID)

	s.db.Where("tunnel_id = ?", tunnel.ID).Delete(&models.Outbound{})

	if err := s.db.Delete(&tunnel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to delete tunnel"))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("Tunnel deleted successfully"))
}

func (s *Server) handleStartTunnel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid tunnel ID"))
		return
	}

	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	var tunnel models.Tunnel
	if err := s.db.Preload("Outbounds").First(&tunnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Tunnel not found"))
		return
	}

	if userRole != "admin" && tunnel.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse("Access denied"))
		return
	}

	if err := s.tunnelMgr.StartTunnel(&tunnel); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(fmt.Sprintf("Failed to start tunnel: %v", err)))
		return
	}

	tunnel.Enabled = true
	s.db.Save(&tunnel)

	c.JSON(http.StatusOK, SuccessMessage("Tunnel started successfully"))
}

func (s *Server) handleStopTunnel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid tunnel ID"))
		return
	}

	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	var tunnel models.Tunnel
	if err := s.db.First(&tunnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Tunnel not found"))
		return
	}

	if userRole != "admin" && tunnel.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse("Access denied"))
		return
	}

	if err := s.tunnelMgr.StopTunnel(tunnel.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(fmt.Sprintf("Failed to stop tunnel: %v", err)))
		return
	}

	tunnel.Enabled = false
	s.db.Save(&tunnel)

	c.JSON(http.StatusOK, SuccessMessage("Tunnel stopped successfully"))
}

func (s *Server) handleRestartTunnel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid tunnel ID"))
		return
	}

	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	var tunnel models.Tunnel
	if err := s.db.Preload("Outbounds").First(&tunnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Tunnel not found"))
		return
	}

	if userRole != "admin" && tunnel.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse("Access denied"))
		return
	}

	if err := s.tunnelMgr.RestartTunnel(&tunnel); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(fmt.Sprintf("Failed to restart tunnel: %v", err)))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("Tunnel restarted successfully"))
}

func (s *Server) handleGetTunnelStats(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid tunnel ID"))
		return
	}

	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	var tunnel models.Tunnel
	if err := s.db.First(&tunnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Tunnel not found"))
		return
	}

	if userRole != "admin" && tunnel.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse("Access denied"))
		return
	}

	upload, download, err := s.tunnelMgr.GetStats(tunnel.ID)
	if err != nil {
		upload = tunnel.UploadBytes
		download = tunnel.DownloadBytes
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"upload":      upload,
		"download":    download,
		"connections": tunnel.Connections,
		"isRunning":   s.tunnelMgr.IsRunning(tunnel.ID),
	}))
}

func (s *Server) handleGetTunnelConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid tunnel ID"))
		return
	}

	userID := c.GetUint("userId")
	userRole, _ := c.Get("role")

	var tunnel models.Tunnel
	if err := s.db.Preload("Outbounds").First(&tunnel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Tunnel not found"))
		return
	}

	if userRole != "admin" && tunnel.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse("Access denied"))
		return
	}

	configJSON, err := s.tunnelMgr.GetConfigJSON(&tunnel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(fmt.Sprintf("Failed to generate config: %v", err)))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"config": configJSON,
	}))
}
