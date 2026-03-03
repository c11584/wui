package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/apps/license-server/internal/api"
	"github.com/your-org/wui/apps/license-server/internal/license"
	"github.com/your-org/wui/apps/license-server/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 初始化数据库
	dbPath := os.Getenv("LICENSE_DB_PATH")
	if dbPath == "" {
		dbPath = "/opt/wui-license/license.db"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&models.License{}, &models.Heartbeat{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database initialized successfully")

	// 初始化服务
	validator := license.NewValidator(db)
	generator := license.NewGenerator(db)
	handler := api.NewHandler(validator, generator)

	// 创建 Gin 路由
	r := gin.Default()

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API 路由
	v1 := r.Group("/api/v1")
	{
		license := v1.Group("/license")
		{
			license.POST("/validate", handler.Validate)
			license.POST("/activate", handler.Activate)
			license.POST("/heartbeat", handler.Heartbeat)
			license.POST("/deactivate", handler.Deactivate)
			license.GET("/info", handler.GetInfo)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务器
	port := os.Getenv("LICENSE_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("License server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
