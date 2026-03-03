package monitor

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
)

type SystemStats struct {
	CPUUsage       float64 `json:"cpuUsage"`
	MemoryUsage    float64 `json:"memoryUsage"`
	DiskUsage      float64 `json:"diskUsage"`
	Uptime         uint64  `json:"uptime"`
	TotalMemory    uint64  `json:"totalMemory"`
	UsedMemory     uint64  `json:"usedMemory"`
	TotalDisk      uint64  `json:"totalDisk"`
	UsedDisk       uint64  `json:"usedDisk"`
	ProcessCount   int     `json:"processCount"`
	GoroutineCount int     `json:"goroutineCount"`
	Timestamp      int64   `json:"timestamp"`
}

type NetworkStats struct {
	BytesSent     uint64  `json:"bytesSent"`
	BytesRecv     uint64  `json:"bytesRecv"`
	PacketsSent   uint64  `json:"packetsSent"`
	PacketsRecv   uint64  `json:"packetsRecv"`
	UploadSpeed   float64 `json:"uploadSpeed"`
	DownloadSpeed float64 `json:"downloadSpeed"`
}

type TunnelStats struct {
	TunnelID      uint       `json:"tunnelId"`
	Connections   int        `json:"connections"`
	UploadBytes   int64      `json:"uploadBytes"`
	DownloadBytes int64      `json:"downloadBytes"`
	UploadSpeed   float64    `json:"uploadSpeed"`
	DownloadSpeed float64    `json:"downloadSpeed"`
	IsRunning     bool       `json:"isRunning"`
	LastConnected *time.Time `json:"lastConnected"`
}

type Monitor struct {
	lastNetworkStats map[string]NetworkStats
	lastUpdateTime   time.Time
}

func NewMonitor() *Monitor {
	return &Monitor{
		lastNetworkStats: make(map[string]NetworkStats),
		lastUpdateTime:   time.Now(),
	}
}

func (m *Monitor) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{
		Timestamp:      time.Now().Unix(),
		GoroutineCount: runtime.NumGoroutine(),
	}

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		stats.CPUUsage = cpuPercent[0]
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		stats.MemoryUsage = memInfo.UsedPercent
		stats.TotalMemory = memInfo.Total
		stats.UsedMemory = memInfo.Used
	}

	diskInfo, err := disk.Usage("/")
	if err == nil {
		stats.DiskUsage = diskInfo.UsedPercent
		stats.TotalDisk = diskInfo.Total
		stats.UsedDisk = diskInfo.Used
	}

	hostInfo, err := host.Info()
	if err == nil {
		stats.Uptime = hostInfo.Uptime
	}

	return stats, nil
}

func (m *Monitor) GetNetworkStats(interfaceName string) (*NetworkStats, error) {
	ioCounters, err := psnet.IOCounters(true)
	if err != nil {
		return nil, err
	}

	var currentStats *NetworkStats
	for _, counter := range ioCounters {
		if counter.Name == interfaceName {
			currentStats = &NetworkStats{
				BytesSent:   counter.BytesSent,
				BytesRecv:   counter.BytesRecv,
				PacketsSent: counter.PacketsSent,
				PacketsRecv: counter.PacketsRecv,
			}
			break
		}
	}

	if currentStats == nil {
		return nil, fmt.Errorf("interface %s not found", interfaceName)
	}

	now := time.Now()
	if lastStats, exists := m.lastNetworkStats[interfaceName]; exists {
		timeDiff := now.Sub(m.lastUpdateTime).Seconds()
		if timeDiff > 0 {
			currentStats.UploadSpeed = float64(currentStats.BytesSent-lastStats.BytesSent) / timeDiff
			currentStats.DownloadSpeed = float64(currentStats.BytesRecv-lastStats.BytesRecv) / timeDiff
		}
	}

	m.lastNetworkStats[interfaceName] = *currentStats
	m.lastUpdateTime = now

	return currentStats, nil
}

func (m *Monitor) GetProcessStats() (int, int64, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	goroutineCount := runtime.NumGoroutine()
	memAlloc := int64(memStats.Alloc)

	return goroutineCount, memAlloc, nil
}
