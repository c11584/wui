package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/apps/license-server/internal/license"
	"github.com/your-org/wui/apps/license-server/internal/models"
)

// Handler handles license API requests
type Handler struct {
	validator *license.Validator
	generator *license.Generator
}

// NewHandler creates a new API handler
func NewHandler(validator *license.Validator, generator *license.Generator) *Handler {
	return &Handler{
		validator: validator,
		generator: generator,
	}
}

// Validate handles POST /api/v1/license/validate
func (h *Handler) Validate(c *gin.Context) {
	var req models.ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.validator.Validate(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Activate handles POST /api/v1/license/activate
func (h *Handler) Activate(c *gin.Context) {
	var req models.ActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.validator.Activate(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Heartbeat handles POST /api/v1/license/heartbeat
func (h *Handler) Heartbeat(c *gin.Context) {
	var req models.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 记录心跳
	if err := h.validator.RecordHeartbeat(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 验证并返回状态
	validateReq := &models.ValidateRequest{
		LicenseKey: req.LicenseKey,
		InstanceID: req.InstanceID,
		IPAddress:  req.IPAddress,
		Domain:     req.Domain,
	}

	resp, err := h.validator.Validate(validateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Deactivate handles POST /api/v1/license/deactivate
func (h *Handler) Deactivate(c *gin.Context) {
	var req struct {
		LicenseKey string `json:"licenseKey" binding:"required"`
		InstanceID string `json:"instanceId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现解绑逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Deactivated successfully"})
}

// GetInfo handles GET /api/v1/license/info
func (h *Handler) GetInfo(c *gin.Context) {
	licenseKey := c.Query("licenseKey")
	if licenseKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "licenseKey is required"})
		return
	}

	validateReq := &models.ValidateRequest{
		LicenseKey: licenseKey,
	}

	resp, err := h.validator.Validate(validateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
