package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/your-org/wui/internal/models"
)

type EmailService struct {
	db     *gorm.DB
	config *SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

type AlertData struct {
	UserName  string
	UserEmail string
	AlertType string
	Message   string
	Timestamp time.Time
	Data      map[string]interface{}
}

func NewEmailService(db *gorm.DB) *EmailService {
	return &EmailService{db: db}
}

func (s *EmailService) LoadConfig() error {
	var settings models.SystemSettings
	result := s.db.First(&settings)
	if result.Error != nil {
		return result.Error
	}

	s.config = &SMTPConfig{
		Host:     settings.SMTPHost,
		Port:     settings.SMTPPort,
		Username: settings.SMTPUsername,
		Password: settings.SMTPPassword,
		From:     settings.SMTPFrom,
		UseTLS:   settings.SMTPTLS,
	}
	return nil
}

func (s *EmailService) Send(to, subject, body string) error {
	if s.config == nil {
		if err := s.LoadConfig(); err != nil {
			return fmt.Errorf("email not configured: %v", err)
		}
	}

	if s.config.Host == "" {
		return fmt.Errorf("smtp not configured")
	}

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0;\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\";\r\n"+
		"\r\n"+
		"%s", s.config.From, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, s.config.From, []string{to}, []byte(msg))
	}

	return smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
}

func (s *EmailService) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: strings.Split(addr, ":")[0]}
		if err = client.StartTLS(config); err != nil {
			return err
		}
	}

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	return w.Close()
}

func (s *EmailService) SendAlertEmail(data AlertData) error {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .content { padding: 30px; }
        .alert-box { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>WUI 通知</h1>
        </div>
        <div class="content">
            <p>尊敬的 <strong>{{.UserName}}</strong>，p>
            <div class="alert-box">
                <strong>{{.AlertType}}</strong><br>
                {{.Message}}
            </div>
            <p>时间：{{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿直接回复。</p>
            <p>© {{.Timestamp.Year}} WUI Proxy Management Panel</p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	subject := fmt.Sprintf("[WUI] %s", data.AlertType)
	return s.Send(data.UserEmail, subject, buf.String())
}

func (s *EmailService) SendTestEmail(to string) error {
	data := AlertData{
		UserName:  "User",
		UserEmail: to,
		AlertType: "测试邮件",
		Message:   "这是一封测试邮件，用于验证邮件服务配置是否正确。",
		Timestamp: time.Now(),
	}
	return s.SendAlertEmail(data)
}
func (s *EmailService) SendWelcomeEmail(user *models.User) error {
	data := AlertData{
		UserName:  user.Username,
		UserEmail: user.Email,
		AlertType: "欢迎注册",
		Message:   "感谢您注册 WUI，开始您的代理管理之旅！",
		Timestamp: time.Now(),
	}
	return s.SendAlertEmail(data)
}

func (s *EmailService) SendPasswordResetEmail(user *models.User, resetLink string) error {
	data := AlertData{
		UserName:  user.Username,
		UserEmail: user.Email,
		AlertType: "密码重置",
		Message:   fmt.Sprintf("请点击以下链接重置您的密码：%s", resetLink),
		Timestamp: time.Now(),
	}
	return s.SendAlertEmail(data)
}

func (s *EmailService) SendAlert(alertType string, userID uint, data map[string]interface{}) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	alertData := AlertData{
		UserName:  user.Username,
		UserEmail: user.Email,
		AlertType: alertType,
		Timestamp: time.Now(),
		Data:      data,
	}

	switch alertType {
	case "traffic_warning":
		if p, ok := data["percent"].(float64); ok {
			alertData.Message = fmt.Sprintf("您的流量使用已达 %.1f%%，请注意控制使用量。", p)
		}
	case "traffic_exhausted":
		alertData.Message = "您的流量已用尽，部分功能将受限。请考虑升级套餐。"
	case "license_expiring":
		if d, ok := data["days"].(float64); ok {
			alertData.Message = fmt.Sprintf("您的许可证将在 %d 天后到期，请及时续费。", int(d))
		}
	case "license_expired":
		alertData.Message = "您的许可证已过期，请续费以继续使用服务。"
	case "order_paid":
		if orderNo, ok := data["order_no"].(string); ok {
			alertData.Message = fmt.Sprintf("订单 %s 已支付成功，感谢您的支持！", orderNo)
		}
	default:
		alertData.Message = "系统通知"
	}

	return s.SendAlertEmail(alertData)
}
