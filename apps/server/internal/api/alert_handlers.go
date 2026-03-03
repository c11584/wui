package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/models"
	"gorm.io/gorm"
)

type AlertManager struct {
	db           *gorm.DB
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	smtpFrom     string
	mu           sync.RWMutex
}

func NewAlertManager(db *gorm.DB) *AlertManager {
	return &AlertManager{db: db}
}

func (a *AlertManager) UpdateSMTP(host string, port int, username, password, from string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.smtpHost = host
	a.smtpPort = port
	a.smtpUsername = username
	a.smtpPassword = password
	a.smtpFrom = from
}

func (a *AlertManager) SendMail(to, subject, body string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.smtpHost == "" {
		return fmt.Errorf("SMTP not configured")
	}

	auth := smtp.PlainAuth("", a.smtpUsername, a.smtpPassword, a.smtpHost)

	msg := bytes.NewBuffer(nil)
	msg.WriteString(fmt.Sprintf("From: %s\r\n", a.smtpFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", a.smtpHost, a.smtpPort)
	return smtp.SendMail(addr, auth, a.smtpFrom, []string{to}, msg.Bytes())
}

func (a *AlertManager) CheckTrafficAlerts() {
	var settings models.SystemSettings
	if err := a.db.First(&settings).Error; err != nil {
		return
	}

	if settings.TrafficAlertPercent <= 0 {
		return
	}

	a.UpdateSMTP(settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername, settings.SMTPPassword, settings.SMTPFrom)

	var users []models.User
	a.db.Where("status = ?", "active").Find(&users)

	for _, user := range users {
		if user.Email == "" {
			continue
		}

		var tunnels []models.Tunnel
		a.db.Where("user_id = ?", user.ID).Find(&tunnels)

		var totalUpload, totalDownload int64
		for _, t := range tunnels {
			totalUpload += t.UploadBytes
			totalDownload += t.DownloadBytes
		}
		totalTraffic := totalUpload + totalDownload

		if user.MaxTraffic > 0 {
			percent := int(float64(totalTraffic) / float64(user.MaxTraffic) * 100)
			if percent >= settings.TrafficAlertPercent {
				subject := "WUI 流量使用告警"
				body := fmt.Sprintf(`
					<h2>流量使用告警</h2>
					<p>您好 %s,</p>
					<p>您的流量使用已达到 %d%%</p>
					<p>已用: %.2f GB / 总量: %.2f GB</p>
					<p>请及时处理以避免服务中断。</p>
				`, user.Username, percent,
					float64(totalTraffic)/1024/1024/1024,
					float64(user.MaxTraffic)/1024/1024/1024)
				a.SendMail(user.Email, subject, body)
			}
		}
	}
}

func (a *AlertManager) CheckLicenseAlerts() {
	var settings models.SystemSettings
	if err := a.db.First(&settings).Error; err != nil {
		return
	}

	if settings.LicenseAlertDays <= 0 {
		return
	}

	a.UpdateSMTP(settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername, settings.SMTPPassword, settings.SMTPFrom)

	var users []models.User
	a.db.Where("status = ? AND license_expire IS NOT NULL", "active").Find(&users)

	now := time.Now()
	alertTime := now.AddDate(0, 0, settings.LicenseAlertDays)

	for _, user := range users {
		if user.Email == "" || user.LicenseExpire == nil {
			continue
		}

		if user.LicenseExpire.Before(alertTime) && user.LicenseExpire.After(now) {
			daysLeft := int(user.LicenseExpire.Sub(now).Hours() / 24)
			subject := "WUI 授权即将到期"
			body := fmt.Sprintf(`
				<h2>授权到期提醒</h2>
				<p>您好 %s,</p>
				<p>您的授权将在 %d 天后到期</p>
				<p>到期时间: %s</p>
				<p>请及时续费以避免服务中断。</p>
			`, user.Username, daysLeft, user.LicenseExpire.Format("2006-01-02"))
			a.SendMail(user.Email, subject, body)
		}
	}
}

func (a *AlertManager) StartScheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			a.CheckTrafficAlerts()
			a.CheckLicenseAlerts()
		}
	}()
}

func (s *Server) handleSendTestEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid email"))
		return
	}

	var settings models.SystemSettings
	if err := s.db.First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("SMTP not configured"))
		return
	}

	auth := smtp.PlainAuth("", settings.SMTPUsername, settings.SMTPPassword, settings.SMTPHost)
	addr := fmt.Sprintf("%s:%d", settings.SMTPHost, settings.SMTPPort)

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: WUI Test Email\r\n\r\nThis is a test email from WUI.", settings.SMTPFrom, req.Email))

	if err := smtp.SendMail(addr, auth, settings.SMTPFrom, []string{req.Email}, msg); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to send email: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("Test email sent successfully"))
}

var globalAlertManager *AlertManager

func InitAlertManager(db *gorm.DB) {
	globalAlertManager = NewAlertManager(db)
	globalAlertManager.StartScheduler()
}
