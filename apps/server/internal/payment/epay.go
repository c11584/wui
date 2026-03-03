package payment

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/your-org/wui/internal/models"
)

type EpayService struct {
	db        *gorm.DB
	apiURL    string
	pid       string
	key       string
	notifyURL string
	returnURL string
}

func NewEpayService(db *gorm.DB, apiURL, pid, key, notifyURL, returnURL string) *EpayService {
	return &EpayService{
		db:        db,
		apiURL:    apiURL,
		pid:       pid,
		key:       key,
		notifyURL: notifyURL,
		returnURL: returnURL,
	}
}

func (s *EpayService) CreatePayment(orderNo string, amount float64, title string, payType string) (string, error) {
	params := url.Values{}
	params.Set("pid", s.pid)
	params.Set("type", payType)
	params.Set("out_trade_no", orderNo)
	params.Set("notify_url", s.notifyURL)
	params.Set("return_url", s.returnURL)
	params.Set("name", title)
	params.Set("money", fmt.Sprintf("%.2f", amount))
	params.Set("sign_type", "MD5")
	params.Set("sign", s.generateSign(params))
	return s.apiURL + "?" + params.Encode(), nil
}

func (s *EpayService) generateSign(params url.Values) string {
	var keys []string
	for k := range params {
		if params.Get(k) != "" && k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, params.Get(k)))
	}
	signStr := strings.Join(pairs, "&")
	signStr += s.key
	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}

func (s *EpayService) VerifyNotify(params url.Values) bool {
	sign := params.Get("sign")
	if sign == "" {
		return false
	}

	receivedSign := sign
	params.Del("sign")
	params.Del("sign_type")

	var keys []string
	for k := range params {
		if params.Get(k) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, params.Get(k)))
	}
	signStr := strings.Join(pairs, "&")
	signStr += s.key

	hash := md5.Sum([]byte(signStr))
	calculatedSign := hex.EncodeToString(hash[:])

	return calculatedSign == receivedSign
}

func (s *EpayService) GetConfig() (*models.PaymentConfig, error) {
	var cfg models.PaymentConfig
	result := s.db.FirstOrCreate(&cfg, models.PaymentConfig{})
	return &cfg, result.Error
}
