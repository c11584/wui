package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	License  LicenseConfig  `json:"license"`
}

type ServerConfig struct {
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	JWTSecret string `json:"jwtSecret"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
}

type LicenseConfig struct {
	KeyLength int `json:"keyLength"`
}

func Load() (*Config, error) {
	configPath := "/opt/wui-admin/config.json"
	if envPath := os.Getenv("WUI_ADMIN_CONFIG"); envPath != "" {
		configPath = envPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Server.JWTSecret == "" {
		cfg.Server.JWTSecret = "wui-admin-secret-change-in-production"
	}
	if cfg.License.KeyLength == 0 {
		cfg.License.KeyLength = 32
	}

	return &cfg, nil
}

func LoadWithDefaults() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Port:      8081,
			Username:  "admin",
			Password:  "admin123",
			JWTSecret: "wui-admin-secret-change-in-production",
		},
		Database: DatabaseConfig{
			Path: "/opt/wui-admin/admin.db",
		},
		License: LicenseConfig{
			KeyLength: 32,
		},
	}

	if envPath := os.Getenv("WUI_ADMIN_CONFIG"); envPath != "" {
		if data, err := os.ReadFile(envPath); err == nil {
			json.Unmarshal(data, cfg)
		}
	}

	return cfg
}
