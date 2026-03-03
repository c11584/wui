package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/your-org/wui/internal/models"
)

type Bot struct {
	db      *gorm.DB
	token   string
	chatID  int64
	enabled bool
	client  *http.Client
}

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		} `json:"chat"`
		Text string `json:"text"`
		Date int    `json:"date"`
	} `json:"message"`
}

type SendMessageRequest struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func NewBot(db *gorm.DB) *Bot {
	return &Bot{
		db: db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (b *Bot) LoadConfig() error {
	var settings models.SystemSettings
	result := b.db.First(&settings)
	if result.Error != nil {
		return result.Error
	}

	if settings.TelegramEnabled {
		b.token = settings.TelegramToken
		b.chatID = settings.TelegramChatID
		b.enabled = true
	}
	return nil
}

func (b *Bot) IsEnabled() bool {
	return b.enabled && b.token != ""
}

func (b *Bot) apiRequest(method string, data interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, method)

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if ok, _ := result["ok"].(bool); !ok {
		return nil, fmt.Errorf("telegram api error: %v", result["description"])
	}

	return result, nil
}

func (b *Bot) SendMessage(text string) error {
	if !b.IsEnabled() {
		return fmt.Errorf("telegram bot not configured")
	}

	_, err := b.apiRequest("sendMessage", SendMessageRequest{
		ChatID:    b.chatID,
		Text:      text,
		ParseMode: "HTML",
	})
	return err
}

func (b *Bot) SendAlert(alertType, message string, data map[string]interface{}) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>🔔 %s</b>\n\n", alertType))
	sb.WriteString(message)
	sb.WriteString(fmt.Sprintf("\n\n⏰ %s", time.Now().Format("2006-01-02 15:04:05")))

	return b.SendMessage(sb.String())
}

func (b *Bot) SendTunnelAlert(tunnelName, status, errMsg string) error {
	emoji := "✅"
	if status == "error" || status == "stopped" {
		emoji = "❌"
	}
	return b.SendAlert(
		"隧道状态变更",
		fmt.Sprintf("%s <b>%s</b>\n状态: %s\n%s", emoji, tunnelName, status, errMsg),
		nil,
	)
}

func (b *Bot) SendTrafficAlert(username string, used, total int64, percent float64) error {
	return b.SendAlert(
		"流量告警",
		fmt.Sprintf("👤 用户: %s\n📊 流量使用: %.1f%%\n💾 已用: %s / %s",
			username,
			percent,
			formatBytes(used),
			formatBytes(total),
		),
		map[string]interface{}{
			"username": username,
			"used":     used,
			"total":    total,
			"percent":  percent,
		},
	)
}

func (b *Bot) SendLicenseAlert(daysLeft int, expiresAt time.Time) error {
	emoji := "⚠️"
	if daysLeft <= 0 {
		emoji = "🚨"
	}
	return b.SendAlert(
		"许可证告警",
		fmt.Sprintf("%s 许可证即将过期\n📅 剩余天数: %d\n🕐 到期时间: %s",
			emoji,
			daysLeft,
			expiresAt.Format("2006-01-02"),
		),
		map[string]interface{}{
			"days_left": daysLeft,
			"expires":   expiresAt,
		},
	)
}

func (b *Bot) SendOrderAlert(orderNo, username string, amount float64, currency string) error {
	return b.SendAlert(
		"新订单通知",
		fmt.Sprintf("💰 订单号: <code>%s</code>\n👤 用户: %s\n💵 金额: %s %.2f",
			orderNo,
			username,
			currency,
			amount,
		),
		map[string]interface{}{
			"order_no": orderNo,
			"username": username,
			"amount":   amount,
			"currency": currency,
		},
	)
}

func (b *Bot) SendSecurityAlert(ip, username, action string) error {
	return b.SendAlert(
		"安全告警",
		fmt.Sprintf("🔒 检测到异常行为\n👤 用户: %s\n🌐 IP: %s\n⚡ 操作: %s",
			username,
			ip,
			action,
		),
		map[string]interface{}{
			"ip":       ip,
			"username": username,
			"action":   action,
		},
	)
}

func (b *Bot) GetUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", b.token, offset)

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("failed to get updates")
	}

	return result.Result, nil
}

func (b *Bot) HandleCommand(update Update) error {
	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID

	if !strings.HasPrefix(text, "/") {
		return nil
	}

	parts := strings.Fields(text)
	cmd := parts[0]

	switch cmd {
	case "/start":
		return b.SendMessage(fmt.Sprintf(
			"👋 欢迎使用 WUI Bot!\n\n您的 Chat ID: <code>%d</code>\n请在设置页面配置此 ID 以接收通知。",
			chatID,
		))
	case "/status":
		return b.handleStatusCommand()
	case "/help":
		return b.SendMessage(
			"<b>WUI Bot 帮助</b>\n\n" +
				"/start - 开始使用\n" +
				"/status - 查看系统状态\n" +
				"/help - 显示帮助信息",
		)
	case "/chatid":
		return b.SendMessage(fmt.Sprintf("您的 Chat ID: <code>%d</code>", chatID))
	}

	return nil
}

func (b *Bot) handleStatusCommand() error {
	var tunnelCount, activeTunnels int64
	var userCount int64

	b.db.Model(&models.Tunnel{}).Count(&tunnelCount)
	b.db.Model(&models.Tunnel{}).Where("enabled = ?", true).Count(&activeTunnels)
	b.db.Model(&models.User{}).Count(&userCount)

	return b.SendMessage(fmt.Sprintf(
		"<b>📊 WUI 系统状态</b>\n\n"+
			"👥 用户总数: %d\n"+
			"🔗 隧道总数: %d\n"+
			"✅ 活跃隧道: %d\n"+
			"⏰ 更新时间: %s",
		userCount,
		tunnelCount,
		activeTunnels,
		time.Now().Format("2006-01-02 15:04:05"),
	))
}

func (b *Bot) StartPolling() {
	if !b.IsEnabled() {
		return
	}

	go func() {
		offset := 0
		for {
			updates, err := b.GetUpdates(offset)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}

			for _, update := range updates {
				offset = update.UpdateID + 1
				if update.Message.Text != "" {
					b.HandleCommand(update)
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

func (b *Bot) SetWebhook(webhookURL string) error {
	_, err := b.apiRequest("setWebhook", map[string]interface{}{
		"url": webhookURL,
	})
	return err
}

func (b *Bot) DeleteWebhook() error {
	_, err := b.apiRequest("deleteWebhook", nil)
	return err
}

func (b *Bot) GetMe() (map[string]interface{}, error) {
	return b.apiRequest("getMe", nil)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
