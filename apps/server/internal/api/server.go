package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/auth"
	"github.com/your-org/wui/internal/config"
	"github.com/your-org/wui/internal/security"
	"github.com/your-org/wui/internal/tunnel"
	"github.com/your-org/wui/internal/websocket"
	"github.com/your-org/wui/internal/xray"
	"gorm.io/gorm"
)

type Server struct {
	config       *config.Config
	db           *gorm.DB
	router       *gin.Engine
	xrayManager  *xray.Manager
	tunnelMgr    *tunnel.Manager
	userHandlers *UserHandlers
	wsHub        *websocket.Hub
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	router := gin.Default()

	xrayMgr := xray.NewManager(cfg.Xray.BinPath, cfg.Xray.ConfigPath)
	tunnelMgr := tunnel.NewManager(xrayMgr)
	userHandlers := NewUserHandlers(db)
	wsHub := websocket.NewHub()

	go wsHub.Run()

	InitAlertManager(db)

	server := &Server{
		config:       cfg,
		db:           db,
		router:       router,
		xrayManager:  xrayMgr,
		tunnelMgr:    tunnelMgr,
		userHandlers: userHandlers,
		wsHub:        wsHub,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	s.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	s.router.Use(func(c *gin.Context) {
		if security.GlobalIPWhitelist.IsEnabled() {
			ip := c.ClientIP()
			if !security.GlobalIPWhitelist.IsAllowed(ip) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Access denied",
				})
				return
			}
		}
		c.Next()
	})

	requireAdminMode := func(c *gin.Context) {
		if s.config.IsAgentMode() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "This feature is not available in agent mode",
			})
			return
		}
		c.Next()
	}

	api := s.router.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", s.userHandlers.Login)
			authGroup.POST("/register", s.userHandlers.Register)
			authGroup.POST("/forgot-password", s.userHandlers.ForgotPassword)
			authGroup.POST("/reset-password", s.userHandlers.ResetPassword)
			authGroup.GET("/captcha", s.handleGetCaptcha)
			authGroup.GET("/login-status", s.handleGetLoginStatus)
		}

		protected := api.Group("")
		protected.Use(auth.Middleware())
		{
			protected.GET("/user", s.userHandlers.GetCurrentUser)
			protected.PUT("/user", s.userHandlers.UpdateCurrentUser)
			protected.GET("/subscription/token", s.handleGetSubscriptionToken)

			api.GET("/subscription", s.handleGetSubscription)

			admin := protected.Group("")
			admin.Use(requireAdminMode)
			admin.Use(auth.RequireRole("admin"))
			{
				admin.GET("/users", s.userHandlers.ListUsers)
				admin.GET("/users/:id", s.userHandlers.GetUser)
				admin.PUT("/users/:id", s.userHandlers.UpdateUser)
				admin.DELETE("/users/:id", s.userHandlers.DeleteUser)
				admin.GET("/audit-logs", s.handleListAuditLogs)

				admin.GET("/settings", s.handleGetSettings)
				admin.PUT("/settings", s.handleUpdateSettings)

				admin.GET("/invite-codes", s.handleListInviteCodes)
				admin.POST("/invite-codes", s.handleCreateInviteCode)
				admin.DELETE("/invite-codes/:id", s.handleDeleteInviteCode)

				admin.GET("/backup", s.handleBackup)
				admin.POST("/restore", s.handleRestore)

				admin.POST("/test-email", s.handleSendTestEmail)

				admin.GET("/licenses", s.handleAdminListLicenses)
				admin.POST("/licenses", s.handleAdminCreateLicense)
				admin.PUT("/licenses/:id", s.handleAdminUpdateLicense)
				admin.DELETE("/licenses/:id", s.handleAdminDeleteLicense)
				admin.POST("/licenses/:id/revoke", s.handleAdminRevokeLicense)
			}

			tunnels := protected.Group("/tunnels")
			{
				tunnels.GET("", s.handleListTunnels)
				tunnels.GET("/:id", s.handleGetTunnel)
				tunnels.POST("", s.handleCreateTunnel)
				tunnels.PUT("/:id", s.handleUpdateTunnel)
				tunnels.DELETE("/:id", s.handleDeleteTunnel)
				tunnels.POST("/:id/start", s.handleStartTunnel)
				tunnels.POST("/:id/stop", s.handleStopTunnel)
				tunnels.POST("/:id/restart", s.handleRestartTunnel)
				tunnels.GET("/:id/stats", s.handleGetTunnelStats)
				tunnels.GET("/:id/config", s.handleGetTunnelConfig)
			}

			system := protected.Group("/system")
			{
				system.GET("/stats", s.handleGetSystemStats)
				system.GET("/version", s.handleGetVersion)
				system.GET("/check-update", s.handleCheckUpdate)
				system.GET("/update/progress", s.handleUpdateProgress)
				system.POST("/update/do", s.handleDoUpdate)
			}

			api.GET("/system/info", s.handleGetSystemInfo)

			protected.POST("/license/activate", s.handleLicenseActivate)
			protected.GET("/license/info", s.handleLicenseInfo)
			protected.POST("/license/deactivate", s.handleLicenseDeactivate)

			// API Tokens
			apiTokens := protected.Group("/api-tokens")
			{
				apiTokens.GET("", s.handleListAPITokens)
				apiTokens.POST("", s.handleCreateAPIToken)
				apiTokens.DELETE("/:id", s.handleDeleteAPIToken)
			}

			// Packages (public for logged-in users)
			packages := protected.Group("/packages")
			{
				packages.GET("", s.handleListPackages)
				packages.GET("/:id", s.handleGetPackage)
			}

			// Orders
			orders := protected.Group("/orders")
			{
				orders.GET("", s.handleListOrders)
				orders.POST("", s.handleCreateOrder)
				orders.GET("/:orderNo", s.handleGetOrder)
			}

			// Payment
			payment := protected.Group("/payment")
			{
				payment.POST("/epay/:orderNo", s.handleEpayPayment)
				payment.POST("/alipay/:orderNo", s.handleAlipayPayment)
			}

			// Payment callbacks
			s.router.POST("/api/payment/epay/notify", s.handleEpayNotify)
			s.router.POST("/api/payment/alipay/notify", s.handleAlipayNotify)

			// Coupons
			protected.GET("/coupons/verify", s.handleVerifyCoupon)

			// Admin commerce routes
			adminCommerce := admin.Group("/commerce")
			{
				// Packages CRUD
				adminCommerce.GET("/packages", s.handleAdminListPackages)
				adminCommerce.POST("/packages", s.handleAdminCreatePackage)
				adminCommerce.PUT("/packages/:id", s.handleAdminUpdatePackage)
				adminCommerce.DELETE("/packages/:id", s.handleAdminDeletePackage)

				// Coupons CRUD
				adminCommerce.GET("/coupons", s.handleAdminListCoupons)
				adminCommerce.POST("/coupons", s.handleAdminCreateCoupon)
				adminCommerce.PUT("/coupons/:id", s.handleAdminUpdateCoupon)
				adminCommerce.DELETE("/coupons/:id", s.handleAdminDeleteCoupon)

				// Orders management
				adminCommerce.GET("/orders", s.handleAdminListOrders)
				adminCommerce.PUT("/orders/:id", s.handleAdminUpdateOrder)

				// Payment config
				adminCommerce.GET("/payment-config", s.handleGetPaymentConfig)
				adminCommerce.PUT("/payment-config", s.handleUpdatePaymentConfig)
			}

			protected.GET("/ws", websocket.HandleWebSocket(s.wsHub))
		}
	}

	s.router.Static("/assets", "./web/assets")
	s.router.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) GetHub() *websocket.Hub {
	return s.wsHub
}
