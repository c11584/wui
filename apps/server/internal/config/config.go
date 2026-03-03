package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Panel    PanelConfig    `json:"panel"`
	Xray     XrayConfig     `json:"xray"`
	Database DatabaseConfig `json:"database"`
	Logs     LogsConfig     `json:"logs"`
	License  LicenseConfig  `json:"license"`
}

type LicenseConfig struct {
	ServerURL       string `json:"serverUrl"`
	GracePeriodDays int    `json:"gracePeriodDays"`
}

type PanelConfig struct {
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type XrayConfig struct {
	BinPath    string `json:"binPath"`
	ConfigPath string `json:"configPath"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
}

type LogsConfig struct {
	Path  string `json:"path"`
	Level string `json:"level"`
}

func Load() (*Config, error) {
	configPath := "/opt/wui/config.json"
	if envPath := os.Getenv("WUI_CONFIG"); envPath != "" {
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

	return &cfg, nil
}
