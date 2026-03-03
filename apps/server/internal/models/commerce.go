package models

import (
	"time"

	"gorm.io/gorm"
)

type Package struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-""`

	Name        string  `gorm:"not null;size:100" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	Price       float64 `gorm:"not null" json:"price"`
	Currency    string  `gorm:"default:'CNY';size:10" json:"currency"`
	Duration    int     `gorm:"not null" json:"duration"` // days
	MaxTunnels  int     `gorm:"default:5" json:"maxTunnels"`
	MaxTraffic  int64   `gorm:"default:107374182400" json:"maxTraffic"` // bytes
	MaxSpeed    int64   `gorm:"default:0" json:"maxSpeed"`              // bytes/s
	Features    string  `gorm:"type:text" json:"features"`              // JSON array
	IsPopular   bool    `gorm:"default:false" json:"isPopular"`
	Enabled     bool    `gorm:"default:true" json:"enabled"`
	SortOrder   int     `gorm:"default:0" json:"sortOrder"`
}

type Order struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	OrderNo   string `gorm:"uniqueIndex;size:32" json:"orderNo"`
	UserID    uint   `gorm:"index" json:"userId"`
	PackageID uint   `gorm:"index" json:"packageId"`

	Amount   float64 `gorm:"not null" json:"amount"`
	Currency string  `gorm:"size:10" json:"currency"`

	Status    string     `gorm:"default:'pending';size:20" json:"status"` // pending, paid, cancelled, refunded
	PayMethod string     `gorm:"size:20" json:"payMethod"`                // alipay, wechat, usdt
	PayTime   *time.Time `json:"payTime"`
	TradeNo   string     `gorm:"size:64" json:"tradeNo"`

	CouponID *uint   `json:"couponId"`
	Discount float64 `gorm:"default:0" json:"discount"`

	ExpireAt *time.Time `json:"expireAt"`
}

type Coupon struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Code        string    `gorm:"uniqueIndex;size:32" json:"code"`
	Discount    float64   `gorm:"not null" json:"discount"` // percentage or fixed amount
	IsPercent   bool      `gorm:"default:true" json:"isPercent"`
	MinAmount   float64   `gorm:"default:0" json:"minAmount"`
	MaxDiscount float64   `gorm:"default:0" json:"maxDiscount"`
	MaxUses     int       `gorm:"default:-1" json:"maxUses"` // -1 = unlimited
	UsedCount   int       `gorm:"default:0" json:"usedCount"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
}

type PaymentConfig struct {
	ID uint `gorm:"primaryKey" json:"id"`

	EpayEnabled   bool   `gorm:"column:epay_enabled;default:false" json:"epayEnabled"`
	EpayAPIURL    string `gorm:"column:epay_api_url;size:255" json:"epayApiUrl"`
	EpayPID       string `gorm:"column:epay_p_id;size:64" json:"epayPid"`
	EpayKey       string `gorm:"column:epay_key;size:64" json:"epayKey"`
	EpayNotifyURL string `gorm:"column:epay_notify_url;size:255" json:"epayNotifyUrl"`
	EpayReturnURL string `gorm:"column:epay_return_url;size:255" json:"epayReturnUrl"`

	AlipayEnabled    bool   `gorm:"column:alipay_enabled;default:false" json:"alipayEnabled"`
	AlipayAppID      string `gorm:"column:alipay_app_id;size:64" json:"alipayAppId"`
	AlipayPrivateKey string `gorm:"column:alipay_private_key;type:text" json:"alipayPrivateKey"`
	AlipayPublicKey  string `gorm:"column:alipay_public_key;type:text" json:"alipayPublicKey"`
	AlipayNotifyURL  string `gorm:"column:alipay_notify_url;size:255" json:"alipayNotifyUrl"`

	WechatEnabled   bool   `gorm:"column:wechat_enabled;default:false" json:"wechatEnabled"`
	WechatAppID     string `gorm:"column:wechat_app_id;size:64" json:"wechatAppId"`
	WechatMchID     string `gorm:"column:wechat_mch_id;size:64" json:"wechatMchId"`
	WechatAPIKey    string `gorm:"column:wechat_api_key;size:64" json:"wechatApiKey"`
	WechatNotifyURL string `gorm:"column:wechat_notify_url;size:255" json:"wechatNotifyUrl"`

	USDTEnabled    bool   `gorm:"column:usdt_enabled;default:false" json:"usdtEnabled"`
	USDTAddress    string `gorm:"column:usdt_address;size:128" json:"usdtAddress"`
	USDTNetwork    string `gorm:"column:usdt_network;size:20" json:"usdtNetwork"`
	USDTMinConfirm int    `gorm:"column:usdt_min_confirm;default:3" json:"usdtMinConfirm"`
}

type APIToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	UserID      uint       `gorm:"index" json:"userId"`
	Name        string     `gorm:"size:100" json:"name"`
	Token       string     `gorm:"uniqueIndex;size:64" json:"token"`
	Permissions string     `gorm:"type:text" json:"permissions"`
	LastUsedAt  *time.Time `json:"lastUsedAt"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	Enabled     bool       `gorm:"default:true" json:"enabled"`
}
