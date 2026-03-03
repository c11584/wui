package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/wui/internal/api"
	"github.com/your-org/wui/internal/config"
	"github.com/your-org/wui/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure directories exist
	os.MkdirAll(cfg.Xray.ConfigPath, 0755)
	os.MkdirAll(cfg.Database.Path[:len(cfg.Database.Path)-len("/wui.db")], 0755)
	os.MkdirAll(cfg.Logs.Path, 0755)

	// Initialize database
	db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Auto migrate
	err = db.AutoMigrate(
		&models.User{}, &models.Tunnel{}, &models.Outbound{}, &models.AuditLog{},
		&models.LicenseCache{}, &models.SystemSettings{}, &models.InviteCode{},
		&models.Package{}, &models.Order{}, &models.Coupon{}, &models.PaymentConfig{},
		&models.APIToken{}, &models.LicenseKey{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Create default admin user if not exists
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.Panel.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		defaultUser := models.User{
			Username:   cfg.Panel.Username,
			Email:      cfg.Panel.Username + "@wui.local",
			Password:   string(hashedPassword),
			Role:       "admin",
			Status:     "active",
			MaxTunnels: 999,
			MaxTraffic: 0,
		}
		db.Create(&defaultUser)
		log.Println("Default admin user created")
	} else {
		// 迁移现有用户：设置默认值
		db.Model(&models.User{}).Where("role = '' OR role IS NULL").Update("role", "admin")
		db.Model(&models.User{}).Where("email = '' OR email IS NULL").Update("email", "admin@wui.local")
		db.Model(&models.User{}).Where("status = '' OR status IS NULL").Update("status", "active")
	}

	// Initialize API server
	server := api.NewServer(cfg, db)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Panel.Port)
		log.Printf("WUI panel started on http://0.0.0.0%s", addr)
		log.Printf("Username: %s", cfg.Panel.Username)
		log.Printf("Password: %s", cfg.Panel.Password)
		if err := server.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down WUI panel...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("Shutdown timeout, forcing exit")
	default:
		log.Println("WUI panel stopped")
	}
}
