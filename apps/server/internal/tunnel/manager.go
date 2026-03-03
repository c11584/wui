package tunnel

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/your-org/wui/internal/models"
	"github.com/your-org/wui/internal/xray"
)

type Manager struct {
	generator   *ConfigGenerator
	xrayManager *xray.Manager
	stats       map[string]*xray.StatsTracker
	mu          sync.RWMutex
}

func NewManager(xrayManager *xray.Manager) *Manager {
	return &Manager{
		generator:   NewConfigGenerator(),
		xrayManager: xrayManager,
		stats:       make(map[string]*xray.StatsTracker),
	}
}

// StartTunnel starts a tunnel
func (m *Manager) StartTunnel(tunnel *models.Tunnel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("%d", tunnel.ID)

	// Generate config
	config, err := m.generator.Generate(tunnel)
	if err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	// Convert to JSON
	configJSON, err := m.generator.ToJSON(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %v", err)
	}

	// Start Xray process
	if err := m.xrayManager.Start(id, configJSON); err != nil {
		return fmt.Errorf("failed to start xray: %v", err)
	}

	// Initialize stats tracker
	m.stats[id] = xray.NewStatsTracker()

	return nil
}

// StopTunnel stops a tunnel
func (m *Manager) StopTunnel(tunnelID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("%d", tunnelID)

	// Stop Xray process
	if err := m.xrayManager.Stop(id); err != nil {
		return fmt.Errorf("failed to stop xray: %v", err)
	}

	// Remove stats tracker
	delete(m.stats, id)

	return nil
}

// RestartTunnel restarts a tunnel
func (m *Manager) RestartTunnel(tunnel *models.Tunnel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("%d", tunnel.ID)

	// Generate config
	config, err := m.generator.Generate(tunnel)
	if err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	// Convert to JSON
	configJSON, err := m.generator.ToJSON(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %v", err)
	}

	// Restart Xray process
	if err := m.xrayManager.Restart(id, configJSON); err != nil {
		return fmt.Errorf("failed to restart xray: %v", err)
	}

	return nil
}

// IsRunning checks if tunnel is running
func (m *Manager) IsRunning(tunnelID uint) bool {
	id := fmt.Sprintf("%d", tunnelID)
	return m.xrayManager.IsRunning(id)
}

// GetStats returns tunnel statistics
func (m *Manager) GetStats(tunnelID uint) (upload int64, download int64, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id := fmt.Sprintf("%d", tunnelID)
	tracker, exists := m.stats[id]
	if !exists {
		return 0, 0, fmt.Errorf("stats tracker not found for tunnel %d", tunnelID)
	}

	upload, download = tracker.GetStats()
	return upload, download, nil
}

// UpdateStats updates tunnel statistics
func (m *Manager) UpdateStats(tunnelID uint, upload, download int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("%d", tunnelID)
	tracker, exists := m.stats[id]
	if !exists {
		return fmt.Errorf("stats tracker not found for tunnel %d", tunnelID)
	}

	tracker.AddUpload(upload)
	tracker.AddDownload(download)

	return nil
}

// GetConfigJSON returns the generated config for a tunnel
func (m *Manager) GetConfigJSON(tunnel *models.Tunnel) (string, error) {
	config, err := m.generator.Generate(tunnel)
	if err != nil {
		return "", fmt.Errorf("failed to generate config: %v", err)
	}

	return m.generator.ToJSON(config)
}

// ValidateConfig validates tunnel configuration
func (m *Manager) ValidateConfig(tunnel *models.Tunnel) error {
	// Validate inbound
	if tunnel.InboundPort < 1 || tunnel.InboundPort > 65535 {
		return fmt.Errorf("invalid inbound port: %d", tunnel.InboundPort)
	}

	if tunnel.InboundProtocol == "" {
		return fmt.Errorf("inbound protocol is required")
	}

	// Validate outbounds (optional - empty means direct connection)
	for i, out := range tunnel.Outbounds {
		if out.Port < 1 || out.Port > 65535 {
			return fmt.Errorf("invalid port for outbound %d: %d", i, out.Port)
		}

		if out.Address == "" {
			return fmt.Errorf("address is required for outbound %d", i)
		}

		if out.Protocol == "" {
			return fmt.Errorf("protocol is required for outbound %d", i)
		}

		// Validate protocol-specific config
		if out.Config != "" {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(out.Config), &config); err != nil {
				return fmt.Errorf("invalid config JSON for outbound %d: %v", i, err)
			}
		}
	}

	return nil
}
