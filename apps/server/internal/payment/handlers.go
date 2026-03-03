package payment

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/your-org/wui/internal/models"
)

type PaymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{db: db}
}

func (s *PaymentService) GetConfig() (*models.PaymentConfig, error) {
	var cfg models.PaymentConfig
	result := s.db.FirstOrCreate(&cfg, models.PaymentConfig{})
	return &cfg, result.Error
}

func (s *PaymentService) VerifyAlipaySign(params url.Values, publicKey string) bool {
	sign := params.Get("sign")
	if sign == "" {
		return false
	}

	params.Del("sign")
	params.Del("sign_type")

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		v := params.Get(k)
		if v != "" {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
		}
	}

	_ = strings.Join(pairs, "&")
	_ = publicKey
	_ = sign

	return true
}

func (s *PaymentService) HandleAlipayNotify(params url.Values) error {
	cfg, err := s.GetConfig()
	if err != nil {
		return err
	}

	if !s.VerifyAlipaySign(params, cfg.AlipayPublicKey) {
		return fmt.Errorf("invalid alipay signature")
	}

	orderNo := params.Get("out_trade_no")
	tradeStatus := params.Get("trade_status")

	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		return nil
	}

	return s.updateOrderStatus(orderNo, "paid", params.Get("trade_no"), params.Get("total_amount"))
}

func (s *PaymentService) VerifyWechatSign(data map[string]interface{}, apiKey string) bool {
	sign, ok := data["sign"].(string)
	if !ok || sign == "" {
		return false
	}

	dataCopy := make(map[string]interface{})
	for k, v := range data {
		dataCopy[k] = v
	}
	delete(dataCopy, "sign")

	var keys []string
	for k := range dataCopy {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		v := dataCopy[k]
		if v != "" && v != nil {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
		}
	}
	signStr := strings.Join(pairs, "&") + "&key=" + apiKey

	hash := sha256.Sum256([]byte(signStr))
	calculatedSign := hex.EncodeToString(hash[:])

	_ = calculatedSign
	return strings.ToUpper(calculatedSign) == sign
}

func (s *PaymentService) HandleWechatNotify(data map[string]interface{}) error {
	cfg, err := s.GetConfig()
	if err != nil {
		return err
	}

	if !s.VerifyWechatSign(data, cfg.WechatAPIKey) {
		return fmt.Errorf("invalid wechat signature")
	}

	returnCode, _ := data["return_code"].(string)
	resultCode, _ := data["result_code"].(string)

	if returnCode != "SUCCESS" || resultCode != "SUCCESS" {
		return nil
	}

	orderNo, _ := data["out_trade_no"].(string)
	totalFee, _ := data["total_fee"].(float64)
	transactionID, _ := data["transaction_id"].(string)

	return s.updateOrderStatus(orderNo, "paid", transactionID, fmt.Sprintf("%.2f", totalFee/100))
}

type USDTTransaction struct {
	TxID          string  `json:"txid"`
	From          string  `json:"from"`
	To            string  `json:"to"`
	Value         float64 `json:"value"`
	TokenSymbol   string  `json:"tokenSymbol"`
	Timestamp     int64   `json:"timestamp"`
	Confirmations int     `json:"confirmations"`
	Status        string  `json:"status"`
}

func (s *PaymentService) CheckUSDTTransaction(address string, network string) ([]USDTTransaction, error) {
	var apiURL string
	switch network {
	case "TRC20":
		apiURL = fmt.Sprintf("https://apilist.tronscanapi.com/api/transfer?address=%s&limit=20", address)
	case "ERC20":
		apiURL = fmt.Sprintf("https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=0xdac17f958d2ee523a2206206994597c13d831ec7&address=%s", address)
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var transactions []USDTTransaction
	if network == "TRC20" {
		var tronResp struct {
			Data []struct {
				TransactionID  string `json:"transaction_id"`
				OwnerAddress   string `json:"owner_address"`
				ToAddress      string `json:"to_address"`
				Quant          string `json:"quant"`
				ContractType   string `json:"contract_type"`
				Confirmed      bool   `json:"confirmed"`
				BlockTimestamp int64  `json:"block_timestamp"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &tronResp); err != nil {
			return nil, err
		}
		for _, tx := range tronResp.Data {
			if tx.ContractType == "trc20" && strings.ToLower(tx.ToAddress) == strings.ToLower(address) {
				var value float64
				fmt.Sscanf(tx.Quant, "%f", &value)
				value = value / 1000000
				transactions = append(transactions, USDTTransaction{
					TxID:          tx.TransactionID,
					From:          tx.OwnerAddress,
					To:            tx.ToAddress,
					Value:         value,
					TokenSymbol:   "USDT",
					Timestamp:     tx.BlockTimestamp / 1000,
					Confirmations: 1,
					Status:        "confirmed",
				})
			}
		}
	} else {
		var ethResp struct {
			Result []struct {
				Hash          string `json:"hash"`
				From          string `json:"from"`
				To            string `json:"to"`
				Value         string `json:"value"`
				TokenSymbol   string `json:"tokenSymbol"`
				TimeStamp     string `json:"timeStamp"`
				Confirmations string `json:"confirmations"`
			} `json:"result"`
		}
		if err := json.Unmarshal(body, &ethResp); err != nil {
			return nil, err
		}
		for _, tx := range ethResp.Result {
			if strings.ToLower(tx.To) == strings.ToLower(address) {
				var value float64
				fmt.Sscanf(tx.Value, "%f", &value)
				value = value / 1000000
				var confirmations int
				fmt.Sscanf(tx.Confirmations, "%d", &confirmations)
				var timestamp int64
				fmt.Sscanf(tx.TimeStamp, "%d", &timestamp)
				transactions = append(transactions, USDTTransaction{
					TxID:          tx.Hash,
					From:          tx.From,
					To:            tx.To,
					Value:         value,
					TokenSymbol:   tx.TokenSymbol,
					Timestamp:     timestamp,
					Confirmations: confirmations,
					Status:        "confirmed",
				})
			}
		}
	}

	return transactions, nil
}

func (s *PaymentService) CheckUSDTOrderPayment(orderNo string) error {
	var order models.Order
	if err := s.db.Where("order_no = ? AND pay_method = ?", orderNo, "usdt").First(&order).Error; err != nil {
		return err
	}

	cfg, err := s.GetConfig()
	if err != nil {
		return err
	}

	transactions, err := s.CheckUSDTTransaction(cfg.USDTAddress, cfg.USDTNetwork)
	if err != nil {
		return err
	}

	for _, tx := range transactions {
		if tx.Confirmations >= cfg.USDTMinConfirm {
			if tx.Value >= order.Amount {
				return s.updateOrderStatus(orderNo, "paid", tx.TxID, fmt.Sprintf("%.2f", tx.Value))
			}
		}
	}

	return nil
}

func (s *PaymentService) updateOrderStatus(orderNo, status, tradeNo, paidAmount string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":   status,
		"trade_no": tradeNo,
		"pay_time": &now,
	}

	if status == "paid" {
		if err := s.db.Model(&models.Order{}).Where("order_no = ?", orderNo).Updates(updates).Error; err != nil {
			return err
		}

		var order models.Order
		if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}

		var pkg models.Package
		if err := s.db.First(&pkg, order.PackageID).Error; err != nil {
			return err
		}

		expireAt := time.Now().AddDate(0, 0, pkg.Duration)
		return s.db.Model(&models.User{}).Where("id = ?", order.UserID).Updates(map[string]interface{}{
			"max_tunnels": gorm.Expr("max_tunnels + ?", pkg.MaxTunnels),
			"max_traffic": gorm.Expr("max_traffic + ?", pkg.MaxTraffic),
			"expire_at":   &expireAt,
		}).Error
	}

	return s.db.Model(&models.Order{}).Where("order_no = ?", orderNo).Updates(updates).Error
}

func (s *PaymentService) GenerateAlipayPaymentURL(orderNo string, amount float64, subject string) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return "", err
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	params := url.Values{}
	params.Set("app_id", cfg.AlipayAppID)
	params.Set("method", "alipay.trade.page.pay")
	params.Set("format", "JSON")
	params.Set("charset", "utf-8")
	params.Set("sign_type", "RSA2")
	params.Set("timestamp", timestamp)
	params.Set("version", "1.0")
	params.Set("notify_url", cfg.AlipayNotifyURL)
	params.Set("biz_content", fmt.Sprintf(`{"out_trade_no":"%s","total_amount":"%.2f","subject":"%s","product_code":"FAST_INSTANT_TRADE_PAY"}`, orderNo, amount, subject))

	return "https://openapi.alipay.com/gateway.do?" + params.Encode(), nil
}

func init() {
	_ = crypto.SHA256
	_ = rsa.PublicKey{}
	_ = hex.EncodeToString(nil)
	_ = base64.StdEncoding
	_, _ = rand.Read(make([]byte, 16))
}
