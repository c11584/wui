package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/auth"
	"github.com/your-org/wui/internal/models"
	"github.com/your-org/wui/internal/security"
	"golang.org/x/crypto/bcrypt"
)

func getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	return c.ClientIP()
}

func (s *Server) handleLogin(c *gin.Context) {
	ip := getClientIP(c)

	if security.GlobalLimiter.IsLocked(ip) {
		remaining := security.GlobalLimiter.GetRemainingLockTime(ip)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Account temporarily locked. Try again in %v", remaining.Round(time.Minute)),
			"locked":  true,
		})
		return
	}

	var req struct {
		Username  string `json:"username" binding:"required"`
		Password  string `json:"password" binding:"required"`
		CaptchaID string `json:"captchaId"`
		Captcha   string `json:"captcha"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	attempts := security.GlobalLimiter.GetAttempts(ip)
	if attempts >= 3 {
		if req.CaptchaID == "" || req.Captcha == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":         false,
				"error":           "Captcha required",
				"captchaRequired": true,
			})
			return
		}
		if !security.GlobalCaptcha.Verify(req.CaptchaID, req.Captcha) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":         false,
				"error":           "Invalid captcha",
				"captchaRequired": true,
			})
			return
		}
	}

	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		security.GlobalLimiter.RecordFailure(ip)
		c.JSON(http.StatusUnauthorized, ErrorResponse("Invalid credentials"))
		return
	}

	if user.Status == "disabled" {
		c.JSON(http.StatusForbidden, ErrorResponse("Account is disabled"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		security.GlobalLimiter.RecordFailure(ip)
		remaining := 5 - security.GlobalLimiter.GetAttempts(ip)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":      false,
			"error":        "Invalid credentials",
			"attemptsLeft": remaining,
		})
		return
	}

	security.GlobalLimiter.RecordSuccess(ip)

	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = ip
	s.db.Save(&user)

	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
			"email":    user.Email,
		},
	}))
}

func (s *Server) handleGetCaptcha(c *gin.Context) {
	id, code := security.GlobalCaptcha.Generate()
	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"captchaId": id,
		"code":      code,
	}))
}

func (s *Server) handleGetLoginStatus(c *gin.Context) {
	ip := getClientIP(c)
	attempts := security.GlobalLimiter.GetAttempts(ip)
	locked := security.GlobalLimiter.IsLocked(ip)

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"attempts":        attempts,
		"locked":          locked,
		"captchaRequired": attempts >= 3,
	}))
}

func (s *Server) handleGetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("userId")

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("User not found"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"id":       user.ID,
		"username": user.Username,
	}))
}

func (s *Server) handleUpdateUser(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	userID, _ := c.Get("userId")

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("User not found"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to hash password"))
		return
	}

	user.Password = string(hashedPassword)

	if err := s.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to update user"))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("User updated successfully"))
}
