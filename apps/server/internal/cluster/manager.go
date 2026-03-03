package cluster

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Node struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	APIToken  string    `json:"-"`
	SecretKey string    `json:"-"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"lastSeen"`
	CPUUsage  float64   `json:"cpuUsage"`
	MemUsage  float64   `json:"memUsage"`
	LoadAvg   string    `json:"loadAvg"`
	Uptime    int64     `json:"uptime"`
	Region    string    `json:"region"`
	Enabled   bool      `json:"enabled"`
	IsLocal   bool      `json:"isLocal"`
}

type NodeManager struct {
	nodes   map[uint]*Node
	localID uint
	mu      sync.RWMutex
	client  *http.Client
}

var GlobalNodeManager *NodeManager

func NewNodeManager(localID uint) *NodeManager {
	mgr := &NodeManager{
		nodes:   make(map[uint]*Node),
		localID: localID,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
	mgr.nodes[localID] = &Node{
		ID:      localID,
		Name:    "Local",
		Status:  "online",
		IsLocal: true,
		Enabled: true,
	}
	GlobalNodeManager = mgr
	return mgr
}

func (m *NodeManager) AddNode(node *Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[node.ID]; exists {
		return fmt.Errorf("node already exists")
	}

	m.nodes[node.ID] = node
	return nil
}

func (m *NodeManager) RemoveNode(id uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == m.localID {
		return
	}
	delete(m.nodes, id)
}

func (m *NodeManager) GetNode(id uint) (*Node, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[id]
	return node, exists
}

func (m *NodeManager) GetAllNodes() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		result = append(result, node)
	}
	return result
}

func (m *NodeManager) GetOnlineNodes() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Node
	for _, node := range m.nodes {
		if node.Status == "online" && node.Enabled {
			result = append(result, node)
		}
	}
	return result
}

func (m *NodeManager) CheckHealth() {
	m.mu.RLock()
	nodes := make([]*Node, 0, len(m.nodes))
	for _, n := range m.nodes {
		nodes = append(nodes, n)
	}
	m.mu.RUnlock()

	for _, node := range nodes {
		if node.IsLocal {
			continue
		}

		go m.checkNodeHealth(node)
	}
}

func (m *NodeManager) checkNodeHealth(node *Node) {
	url := fmt.Sprintf("http://%s:%d/api/health", node.Address, node.Port)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+node.APIToken)
	req.Header.Set("X-Node-Signature", m.signRequest(node.SecretKey, fmt.Sprintf("%d", time.Now().Unix())))

	resp, err := m.client.Do(req)
	if err != nil {
		m.updateNodeStatus(node.ID, "offline")
		return
	}
	defer resp.Body.Close()

	var health struct {
		CPUUsage float64 `json:"cpuUsage"`
		MemUsage float64 `json:"memUsage"`
		LoadAvg  string  `json:"loadAvg"`
		Uptime   int64   `json:"uptime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		m.updateNodeStatus(node.ID, "error")
		return
	}

	m.mu.Lock()
	if n, exists := m.nodes[node.ID]; exists {
		n.Status = "online"
		n.LastSeen = time.Now()
		n.CPUUsage = health.CPUUsage
		n.MemUsage = health.MemUsage
		n.LoadAvg = health.LoadAvg
		n.Uptime = health.Uptime
	}
	m.mu.Unlock()
}

func (m *NodeManager) updateNodeStatus(id uint, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if node, exists := m.nodes[id]; exists {
		node.Status = status
		node.LastSeen = time.Now()
	}
}

func (m *NodeManager) signRequest(secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (m *NodeManager) ForwardRequest(nodeID uint, path string, method string, body []byte) ([]byte, error) {
	m.mu.RLock()
	node, exists := m.nodes[nodeID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node not found")
	}

	if node.Status != "online" {
		return nil, fmt.Errorf("node offline")
	}

	url := fmt.Sprintf("http://%s:%d%s", node.Address, node.Port, path)

	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, url, nil)
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}

	req.Header.Set("Authorization", "Bearer "+node.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Signature", m.signRequest(node.SecretKey, fmt.Sprintf("%d", time.Now().Unix())))

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []byte
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (m *NodeManager) StartHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m.CheckHealth()
		}
	}()
}

func (m *NodeManager) SyncTunnelConfig(nodeID uint, tunnelConfig []byte) error {
	m.mu.RLock()
	node, exists := m.nodes[nodeID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("node not found")
	}

	url := fmt.Sprintf("http://%s:%d/api/sync/tunnel", node.Address, node.Port)

	req, err := http.NewRequest("POST", url, bytes.NewReader(tunnelConfig))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+node.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Signature", m.signRequest(node.SecretKey, string(tunnelConfig)))

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sync failed: %d", resp.StatusCode)
	}

	return nil
}

func (m *NodeManager) SyncUserQuota(nodeID uint, userID uint, maxTunnels int, maxTraffic int64) error {
	m.mu.RLock()
	node, exists := m.nodes[nodeID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("node not found")
	}

	payload := map[string]interface{}{
		"userId":     userID,
		"maxTunnels": maxTunnels,
		"maxTraffic": maxTraffic,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("http://%s:%d/api/sync/quota", node.Address, node.Port)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+node.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Signature", m.signRequest(node.SecretKey, string(body)))

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sync failed: %d", resp.StatusCode)
	}

	return nil
}

func (m *NodeManager) BroadcastConfig(tunnelConfig []byte) error {
	nodes := m.GetOnlineNodes()
	var errors []error

	for _, node := range nodes {
		if !node.IsLocal {
			if err := m.SyncTunnelConfig(node.ID, tunnelConfig); err != nil {
				errors = append(errors, fmt.Errorf("node %d: %v", node.ID, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync errors: %v", errors)
	}
	return nil
}

func (m *NodeManager) GetClusterStats() map[string]interface{} {
	nodes := m.GetAllNodes()
	online := 0
	totalCPU := 0.0
	totalMem := 0.0

	for _, node := range nodes {
		if node.Status == "online" {
			online++
			totalCPU += node.CPUUsage
			totalMem += node.MemUsage
		}
	}

	return map[string]interface{}{
		"totalNodes":   len(nodes),
		"onlineNodes":  online,
		"offlineNodes": len(nodes) - online,
		"avgCPU":       totalCPU / float64(max(online, 1)),
		"avgMem":       totalMem / float64(max(online, 1)),
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
