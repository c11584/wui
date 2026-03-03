package stats

import (
	"sync"
	"time"
)

type DataPoint struct {
	Timestamp int64 `json:"timestamp"`
	Upload    int64 `json:"upload"`
	Download  int64 `json:"download"`
}

type ConnectionInfo struct {
	TunnelID    uint      `json:"tunnelId"`
	ClientIP    string    `json:"clientIp"`
	ConnectedAt time.Time `json:"connectedAt"`
	BytesUp     int64     `json:"bytesUp"`
	BytesDown   int64     `json:"bytesDown"`
}

type TunnelStats struct {
	TunnelID    uint  `json:"tunnelId"`
	IsRunning   bool  `json:"isRunning"`
	Upload      int64 `json:"upload"`
	Download    int64 `json:"download"`
	Connections int   `json:"connections"`
}

type Collector struct {
	stats   map[uint][]DataPoint
	conns   map[uint][]ConnectionInfo
	history map[uint][]DataPoint
	mu      sync.RWMutex
}

func NewCollector() *Collector {
	return &Collector{
		stats:   make(map[uint][]DataPoint),
		conns:   make(map[uint][]ConnectionInfo),
		history: make(map[uint][]DataPoint),
	}
}

func (c *Collector) RecordConnection(tunnelID uint, info ConnectionInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conns[tunnelID] = append(c.conns[tunnelID], info)
}

func (c *Collector) RecordTraffic(tunnelID uint, upload, download int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().Unix()
	point := DataPoint{
		Timestamp: now,
		Upload:    upload,
		Download:  download,
	}

	c.stats[tunnelID] = append(c.stats[tunnelID], point)

	if len(c.stats[tunnelID]) > 288 {
		c.stats[tunnelID] = c.stats[tunnelID][1:]
	}

	c.history[tunnelID] = append(c.history[tunnelID], point)
	if len(c.history[tunnelID]) > 1440 {
		c.history[tunnelID] = c.history[tunnelID][len(c.history[tunnelID])-1440:]
	}
}

func (c *Collector) GetActiveConnections(tunnelID uint) []ConnectionInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	conns := c.conns[tunnelID]
	c.conns[tunnelID] = nil
	return conns
}

func (c *Collector) GetStats(tunnelID uint) *TunnelStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats[tunnelID]
	if len(stats) == 0 {
		return &TunnelStats{TunnelID: tunnelID, IsRunning: false}
	}

	var upload, download int64
	for _, p := range stats {
		upload += p.Upload
		download += p.Download
	}

	return &TunnelStats{
		TunnelID:    tunnelID,
		IsRunning:   true,
		Upload:      upload,
		Download:    download,
		Connections: len(c.conns[tunnelID]),
	}
}

func (c *Collector) GetHistory(tunnelID uint, points int) []DataPoint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	history := c.history[tunnelID]
	if len(history) == 0 {
		return nil
	}

	if points <= 0 || points > len(history) {
		return history
	}
	return history[len(history)-points:]
}

var GlobalCollector = NewCollector()
