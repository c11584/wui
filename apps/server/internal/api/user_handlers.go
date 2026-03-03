package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/auth"
	"github.com/your-org/wui/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandlers struct {
	db *gorm.DB
}

func NewUserHandlers(db *gorm.DB) *UserHandlers {
	return &UserHandlers{db: db}
}

type RegisterRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=32"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"inviteCode"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Email      string `json:"email" binding:"omitempty,email"`
	Password   string `json:"password" binding:"omitempty,min=6"`
	MaxTunnels *int   `json:"maxTunnels"`
	MaxTraffic *int64 `json:"maxTraffic"`
	Status     string `json:"status" binding:"omitempty,oneof=active suspended"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandlers) Register(c *gin.Context) {
	var settings models.SystemSettings
	h.db.First(&settings)

	if !settings.RegistrationEnabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Registration is disabled"})
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if settings.InviteOnly {
		if req.InviteCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invite code required"})
			return
		}

		var invite models.InviteCode
		if err := h.db.Where("code = ?", req.InviteCode).First(&invite).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invite code"})
			return
		}

		if invite.ExpiresAt != nil && time.Now().After(*invite.ExpiresAt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invite code expired"})
			return
		}

		if invite.UsedCount >= invite.MaxUses {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invite code has reached maximum uses"})
			return
		}

		h.db.Model(&invite).Update("used_count", invite.UsedCount+1)
	}

	var existingUser models.User
	if h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   string(hashedPassword),
		Role:       "user",
		Status:     "active",
		MaxTunnels: 5,
		MaxTraffic: 107374182400,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	h.logAudit(c, user.ID, user.Username, "register", "users", int(user.ID), "User registered", true)

	token, _ := auth.GenerateToken(user.ID, user.Username, user.Role)

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *UserHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is " + user.Status})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	now := time.Now()
	ip := c.ClientIP()
	h.db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": ip,
	})

	token, _ := auth.GenerateToken(user.ID, user.Username, user.Role)

	h.logAudit(c, user.ID, user.Username, "login", "users", int(user.ID), "User logged in", true)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *UserHandlers) GetCurrentUser(c *gin.Context) {
	userID := c.GetUint("userId")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               user.ID,
		"username":         user.Username,
		"email":            user.Email,
		"role":             user.Role,
		"status":           user.Status,
		"maxTunnels":       user.MaxTunnels,
		"maxTraffic":       user.MaxTraffic,
		"twoFactorEnabled": user.TwoFactorEnabled,
		"licenseKey":       user.LicenseKey,
		"licenseExpire":    user.LicenseExpire,
	})
}

func (h *UserHandlers) UpdateCurrentUser(c *gin.Context) {
	userID := c.GetUint("userId")

	var req struct {
		Email    string `json:"email" binding:"omitempty,email"`
		Password string `json:"password" binding:"omitempty,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		updates["password"] = string(hashedPassword)
	}

	if len(updates) > 0 {
		if err := h.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	h.logAudit(c, userID, "", "update_profile", "users", int(userID), "User updated profile", true)

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

func (h *UserHandlers) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "If email exists, reset link sent"})
		return
	}

	token := generateResetToken()
	exp := time.Now().Add(1 * time.Hour)

	h.db.Model(&user).Updates(map[string]interface{}{
		"reset_token":     token,
		"reset_token_exp": exp,
	})

	log.Printf("[DEV] Password reset token for %s: %s", user.Email, token)

	h.logAudit(c, user.ID, user.Username, "forgot_password", "users", int(user.ID), "Password reset requested", true)

	c.JSON(http.StatusOK, gin.H{"message": "If email exists, reset link sent"})
}

func (h *UserHandlers) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("reset_token = ? AND reset_token_exp > ?", req.Token, time.Now()).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	h.db.Model(&user).Updates(map[string]interface{}{
		"password":        string(hashedPassword),
		"reset_token":     "",
		"reset_token_exp": nil,
	})

	h.logAudit(c, user.ID, user.Username, "reset_password", "users", int(user.ID), "Password reset completed", true)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

func (h *UserHandlers) ListUsers(c *gin.Context) {
	var users []models.User
	h.db.Select("id, username, email, role, status, max_tunnels, max_traffic, created_at, last_login_at").Find(&users)

	c.JSON(http.StatusOK, users)
}

func (h *UserHandlers) GetUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := h.db.Select("id, username, email, role, status, max_tunnels, max_traffic, created_at, last_login_at, license_key, license_expire").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandlers) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		updates["password"] = string(hashedPassword)
	}
	if req.MaxTunnels != nil {
		updates["max_tunnels"] = *req.MaxTunnels
	}
	if req.MaxTraffic != nil {
		updates["max_traffic"] = *req.MaxTraffic
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := h.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	adminID := c.GetUint("userId")
	h.logAudit(c, adminID, "", "update_user", "users", parseInt(userID), "Admin updated user", true)

	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func (h *UserHandlers) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.db.Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	adminID := c.GetUint("userId")
	h.logAudit(c, adminID, "", "delete_user", "users", parseInt(userID), "Admin deleted user", true)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func (h *UserHandlers) logAudit(c *gin.Context, userID uint, username, action, resource string, resourceID int, detail string, success bool) {
	if username == "" {
		username = c.GetString("username")
	}

	detailJSON, _ := json.Marshal(map[string]interface{}{"message": detail})

	log := models.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: uint(resourceID),
		Detail:     string(detailJSON),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Success:    success,
	}

	h.db.Create(&log)
}

func generateResetToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}
