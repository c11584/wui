package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/your-org/wui/internal/models"
	"github.com/your-org/wui/internal/payment"
)

func (s *Server) handleEpayPayment(c *gin.Context) {
	userID := c.GetUint("userId")
	orderNo := c.Param("orderNo")

	var order models.Order
	if err := s.db.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Order not found"))
		return
	}

	if order.Status != "pending" {
		c.JSON(http.StatusBadRequest, ErrorResponse("Order already paid or cancelled"))
		return
	}

	var cfg models.PaymentConfig
	s.db.FirstOrCreate(&cfg, models.PaymentConfig{})

	if !cfg.EpayEnabled {
		c.JSON(http.StatusBadRequest, ErrorResponse("Epay payment not enabled"))
		return
	}

	var pkg models.Package
	if err := s.db.First(&pkg, order.PackageID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Package not found"))
		return
	}

	epaySvc := payment.NewEpayService(s.db, cfg.EpayAPIURL, cfg.EpayPID, cfg.EpayKey, cfg.EpayNotifyURL, cfg.EpayReturnURL)
	payURL, err := epaySvc.CreatePayment(orderNo, order.Amount, pkg.Name, "alipay")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create payment"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"payUrl": payURL,
	}))
}

func (s *Server) handleEpayNotify(c *gin.Context) {
	var cfg models.PaymentConfig
	s.db.FirstOrCreate(&cfg, models.PaymentConfig{})

	epaySvc := payment.NewEpayService(s.db, cfg.EpayAPIURL, cfg.EpayPID, cfg.EpayKey, cfg.EpayNotifyURL, cfg.EpayReturnURL)

	if !epaySvc.VerifyNotify(c.Request.URL.Query()) {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	orderNo := c.Query("out_trade_no")
	tradeStatus := c.Query("trade_status")

	if tradeStatus != "TRADE_SUCCESS" {
		c.String(http.StatusOK, "success")
		return
	}

	var order models.Order
	if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		c.String(http.StatusOK, "success")
		return
	}

	if order.Status == "paid" {
		c.String(http.StatusOK, "success")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":   "paid",
		"trade_no": c.Query("trade_no"),
		"pay_time": &now,
	}

	if err := s.db.Model(&order).Updates(updates).Error; err != nil {
		c.String(http.StatusOK, "success")
		return
	}

	var pkg models.Package
	if err := s.db.First(&pkg, order.PackageID).Error; err == nil {
		expireAt := time.Now().AddDate(0, 0, pkg.Duration)
		s.db.Model(&models.User{}).Where("id = ?", order.UserID).Updates(map[string]interface{}{
			"max_tunnels": gormExpr("max_tunnels + ?", pkg.MaxTunnels),
			"max_traffic": gormExpr("max_traffic + ?", pkg.MaxTraffic),
			"expire_at":   &expireAt,
		})
	}

	c.String(http.StatusOK, "success")
}

func (s *Server) handleAlipayPayment(c *gin.Context) {
	userID := c.GetUint("userId")
	orderNo := c.Param("orderNo")

	var order models.Order
	if err := s.db.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Order not found"))
		return
	}

	if order.Status != "pending" {
		c.JSON(http.StatusBadRequest, ErrorResponse("Order already paid or cancelled"))
		return
	}

	var cfg models.PaymentConfig
	s.db.FirstOrCreate(&cfg, models.PaymentConfig{})

	if !cfg.AlipayEnabled {
		c.JSON(http.StatusBadRequest, ErrorResponse("Alipay payment not enabled"))
		return
	}

	var pkg models.Package
	if err := s.db.First(&pkg, order.PackageID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Package not found"))
		return
	}

	payURL, err := payment.NewPaymentService(s.db).GenerateAlipayPaymentURL(orderNo, order.Amount, pkg.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("Failed to create payment"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"payUrl": payURL,
	}))
}

func (s *Server) handleAlipayNotify(c *gin.Context) {
	paySvc := payment.NewPaymentService(s.db)

	if err := paySvc.HandleAlipayNotify(c.Request.URL.Query()); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

func gormExpr(expr string, value interface{}) clause.Expr {
	return clause.Expr{SQL: expr, Vars: []interface{}{value}}
}
