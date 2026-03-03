package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/models"
)

func (s *Server) handleListAPITokens(c *gin.Context) {
	userID := c.GetUint("userId")

	var tokens []models.APIToken
	s.db.Where("user_id = ?", userID).Find(&tokens)

	c.JSON(http.StatusOK, SuccessResponse(tokens))
}

func (s *Server) handleCreateAPIToken(c *gin.Context) {
	userID := c.GetUint("userId")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Permissions string `json:"permissions"`
		ExpiresIn   *int   `json:"expiresIn"` // hours
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	bytes := make([]byte, 32)
	rand.Read(bytes)
	token := "wui_" + hex.EncodeToString(bytes)

	apiToken := models.APIToken{
		UserID:      userID,
		Name:        req.Name,
		Token:       token,
		Permissions: req.Permissions,
		Enabled:     true,
	}

	if req.ExpiresIn != nil {
		exp := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		apiToken.ExpiresAt = &exp
	}

	if err := s.db.Create(&apiToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create token"))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(apiToken))
}

func (s *Server) handleDeleteAPIToken(c *gin.Context) {
	userID := c.GetUint("userId")
	id := c.Param("id")

	if err := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.APIToken{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to delete token"))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("Token deleted"))
}

func (s *Server) handleListPackages(c *gin.Context) {
	var packages []models.Package
	s.db.Where("enabled = ?", true).Order("sort_order").Find(&packages)

	c.JSON(http.StatusOK, SuccessResponse(packages))
}

func (s *Server) handleGetPackage(c *gin.Context) {
	id := c.Param("id")

	var pkg models.Package
	if err := s.db.First(&pkg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Package not found"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(pkg))
}

func (s *Server) handleCreateOrder(c *gin.Context) {
	userID := c.GetUint("userId")

	var req struct {
		PackageID  uint   `json:"packageId" binding:"required"`
		CouponCode string `json:"couponCode"`
		PayMethod  string `json:"payMethod" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	var pkg models.Package
	if err := s.db.First(&pkg, req.PackageID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Package not found"))
		return
	}

	order := models.Order{
		OrderNo:   GenerateOrderNo(),
		UserID:    userID,
		PackageID: pkg.ID,
		Amount:    pkg.Price,
		Currency:  pkg.Currency,
		Status:    "pending",
		PayMethod: req.PayMethod,
	}

	if req.CouponCode != "" {
		var coupon models.Coupon
		if err := s.db.Where("code = ? AND enabled = ?", req.CouponCode, true).First(&coupon).Error; err == nil {
			if time.Now().After(coupon.EndTime) || time.Now().Before(coupon.StartTime) {
				c.JSON(http.StatusBadRequest, ErrorResponse("Coupon expired or not yet valid"))
				return
			}
			if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
				c.JSON(http.StatusBadRequest, ErrorResponse("Coupon usage limit reached"))
				return
			}
			if order.Amount >= coupon.MinAmount {
				order.CouponID = &coupon.ID
				if coupon.IsPercent {
					order.Discount = order.Amount * coupon.Discount / 100
					if coupon.MaxDiscount > 0 && order.Discount > coupon.MaxDiscount {
						order.Discount = coupon.MaxDiscount
					}
				} else {
					order.Discount = coupon.Discount
				}
				order.Amount -= order.Discount
			}
		}
	}

	if err := s.db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create order"))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(order))
}

func (s *Server) handleGetOrder(c *gin.Context) {
	userID := c.GetUint("userId")
	orderNo := c.Param("orderNo")

	var order models.Order
	if err := s.db.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Order not found"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(order))
}

func (s *Server) handleListOrders(c *gin.Context) {
	userID := c.GetUint("userId")

	var orders []models.Order
	s.db.Where("user_id = ?", userID).Order("created_at desc").Limit(50).Find(&orders)

	c.JSON(http.StatusOK, SuccessResponse(orders))
}

func (s *Server) handleVerifyCoupon(c *gin.Context) {
	code := c.Query("code")
	amount := 0.0
	fmt.Sscanf(c.Query("amount"), "%f", &amount)

	if code == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("Coupon code required"))
		return
	}

	var coupon models.Coupon
	if err := s.db.Where("code = ? AND enabled = ?", code, true).First(&coupon).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Invalid coupon"))
		return
	}

	if time.Now().After(coupon.EndTime) {
		c.JSON(http.StatusBadRequest, ErrorResponse("Coupon expired"))
		return
	}

	if time.Now().Before(coupon.StartTime) {
		c.JSON(http.StatusBadRequest, ErrorResponse("Coupon not yet valid"))
		return
	}

	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		c.JSON(http.StatusBadRequest, ErrorResponse("Coupon usage limit reached"))
		return
	}

	if amount < coupon.MinAmount {
		c.JSON(http.StatusBadRequest, ErrorResponse(fmt.Sprintf("Minimum order amount: %.2f", coupon.MinAmount)))
		return
	}

	var discount float64
	if coupon.IsPercent {
		discount = amount * coupon.Discount / 100
		if coupon.MaxDiscount > 0 && discount > coupon.MaxDiscount {
			discount = coupon.MaxDiscount
		}
	} else {
		discount = coupon.Discount
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"valid":     true,
		"discount":  discount,
		"isPercent": coupon.IsPercent,
	}))
}

func GenerateOrderNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("WUI%d%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

func (s *Server) handleAdminListPackages(c *gin.Context) {
	var packages []models.Package
	s.db.Order("sort_order").Find(&packages)
	c.JSON(http.StatusOK, SuccessResponse(packages))
}

func (s *Server) handleAdminCreatePackage(c *gin.Context) {
	var req models.Package
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	if err := s.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create package"))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(req))
}

func (s *Server) handleAdminUpdatePackage(c *gin.Context) {
	id := c.Param("id")

	var pkg models.Package
	if err := s.db.First(&pkg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Package not found"))
		return
	}

	var req models.Package
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	req.ID = pkg.ID
	if err := s.db.Save(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to update package"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(req))
}

func (s *Server) handleAdminDeletePackage(c *gin.Context) {
	id := c.Param("id")

	if err := s.db.Delete(&models.Package{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to delete package"))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("Package deleted"))
}

func (s *Server) handleAdminListCoupons(c *gin.Context) {
	var coupons []models.Coupon
	s.db.Order("created_at desc").Find(&coupons)
	c.JSON(http.StatusOK, SuccessResponse(coupons))
}

func (s *Server) handleAdminCreateCoupon(c *gin.Context) {
	var req models.Coupon
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	if err := s.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create coupon"))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(req))
}

func (s *Server) handleAdminUpdateCoupon(c *gin.Context) {
	id := c.Param("id")

	var coupon models.Coupon
	if err := s.db.First(&coupon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Coupon not found"))
		return
	}

	var req models.Coupon
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	req.ID = coupon.ID
	if err := s.db.Save(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to update coupon"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(req))
}

func (s *Server) handleAdminDeleteCoupon(c *gin.Context) {
	id := c.Param("id")

	if err := s.db.Delete(&models.Coupon{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to delete coupon"))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("Coupon deleted"))
}

func (s *Server) handleAdminListOrders(c *gin.Context) {
	var orders []models.Order
	s.db.Preload("User").Preload("Package").Order("created_at desc").Limit(100).Find(&orders)
	c.JSON(http.StatusOK, SuccessResponse(orders))
}

func (s *Server) handleAdminUpdateOrder(c *gin.Context) {
	id := c.Param("id")

	var order models.Order
	if err := s.db.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Order not found"))
		return
	}

	var req struct {
		Status        string  `json:"status"`
		PaidAmount    float64 `json:"paidAmount"`
		TransactionID string  `json:"transactionId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	updates := map[string]interface{}{}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.PaidAmount > 0 {
		updates["paid_amount"] = req.PaidAmount
	}
	if req.TransactionID != "" {
		updates["transaction_id"] = req.TransactionID
	}

	if err := s.db.Model(&order).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to update order"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(order))
}

func (s *Server) handleGetPaymentConfig(c *gin.Context) {
	var cfg models.PaymentConfig
	s.db.FirstOrCreate(&cfg, models.PaymentConfig{})

	epayKey := cfg.EpayKey
	if len(epayKey) > 4 {
		epayKey = "****" + epayKey[len(epayKey)-4:]
	}

	result := gin.H{
		"epay": gin.H{
			"enabled":   cfg.EpayEnabled,
			"apiUrl":    cfg.EpayAPIURL,
			"pid":       cfg.EpayPID,
			"key":       epayKey,
			"notifyUrl": cfg.EpayNotifyURL,
			"returnUrl": cfg.EpayReturnURL,
		},
		"alipay": gin.H{
			"enabled":   cfg.AlipayEnabled,
			"appId":     cfg.AlipayAppID,
			"notifyUrl": cfg.AlipayNotifyURL,
		},
		"wechat": gin.H{
			"enabled":   cfg.WechatEnabled,
			"appId":     cfg.WechatAppID,
			"mchId":     cfg.WechatMchID,
			"notifyUrl": cfg.WechatNotifyURL,
		},
		"usdt": gin.H{
			"enabled":    cfg.USDTEnabled,
			"address":    cfg.USDTAddress,
			"network":    cfg.USDTNetwork,
			"minConfirm": cfg.USDTMinConfirm,
		},
	}

	c.JSON(http.StatusOK, SuccessResponse(result))
}

func (s *Server) handleUpdatePaymentConfig(c *gin.Context) {
	var cfg models.PaymentConfig
	s.db.FirstOrCreate(&cfg, models.PaymentConfig{})

	var req map[string]map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request"))
		return
	}

	if epay, ok := req["epay"]; ok {
		if enabled, ok := epay["enabled"].(bool); ok {
			cfg.EpayEnabled = enabled
		}
		if apiUrl, ok := epay["apiUrl"].(string); ok {
			cfg.EpayAPIURL = apiUrl
		}
		if pid, ok := epay["pid"].(string); ok {
			cfg.EpayPID = pid
		}
		if key, ok := epay["key"].(string); ok {
			cfg.EpayKey = key
		}
		if notifyUrl, ok := epay["notifyUrl"].(string); ok {
			cfg.EpayNotifyURL = notifyUrl
		}
		if returnUrl, ok := epay["returnUrl"].(string); ok {
			cfg.EpayReturnURL = returnUrl
		}
	}

	if alipay, ok := req["alipay"]; ok {
		if enabled, ok := alipay["enabled"].(bool); ok {
			cfg.AlipayEnabled = enabled
		}
		if appId, ok := alipay["appId"].(string); ok {
			cfg.AlipayAppID = appId
		}
		if privateKey, ok := alipay["privateKey"].(string); ok {
			cfg.AlipayPrivateKey = privateKey
		}
		if publicKey, ok := alipay["publicKey"].(string); ok {
			cfg.AlipayPublicKey = publicKey
		}
		if notifyUrl, ok := alipay["notifyUrl"].(string); ok {
			cfg.AlipayNotifyURL = notifyUrl
		}
	}

	if wechat, ok := req["wechat"]; ok {
		if enabled, ok := wechat["enabled"].(bool); ok {
			cfg.WechatEnabled = enabled
		}
		if appId, ok := wechat["appId"].(string); ok {
			cfg.WechatAppID = appId
		}
		if mchId, ok := wechat["mchId"].(string); ok {
			cfg.WechatMchID = mchId
		}
		if apiKey, ok := wechat["apiKey"].(string); ok {
			cfg.WechatAPIKey = apiKey
		}
		if notifyUrl, ok := wechat["notifyUrl"].(string); ok {
			cfg.WechatNotifyURL = notifyUrl
		}
	}

	if usdt, ok := req["usdt"]; ok {
		if enabled, ok := usdt["enabled"].(bool); ok {
			cfg.USDTEnabled = enabled
		}
		if address, ok := usdt["address"].(string); ok {
			cfg.USDTAddress = address
		}
		if network, ok := usdt["network"].(string); ok {
			cfg.USDTNetwork = network
		}
		if minConfirm, ok := usdt["minConfirm"].(float64); ok {
			cfg.USDTMinConfirm = int(minConfirm)
		}
	}

	if err := s.db.Save(&cfg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to update payment config"))
		return
	}

	c.JSON(http.StatusOK, SuccessMessage("Payment config updated"))
}
