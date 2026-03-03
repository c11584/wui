package license

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ValidateResponse struct {
	Valid      bool   `json:"valid"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	Plan       string `json:"plan"`
	MaxTunnels int    `json:"maxTunnels"`
	MaxUsers   int    `json:"maxUsers"`
	MaxTraffic int64  `json:"maxTraffic"`
	Features   string `json:"features"`
	ExpiresAt  string `json:"expiresAt"`
}

type Client struct {
	serverURL string
	client    *http.Client
}

func NewClient(serverURL string) *Client {
	if serverURL == "" {
		serverURL = "http://localhost:8090"
	}
	return &Client{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Validate(licenseKey, instanceID string) (*ValidateResponse, error) {
	reqBody := map[string]string{
		"licenseKey": licenseKey,
		"instanceId": instanceID,
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := c.client.Post(c.serverURL+"/api/v1/license/validate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("license server unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ValidateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &result, nil
}

func (c *Client) Activate(licenseKey, instanceID string) (*ValidateResponse, error) {
	reqBody := map[string]string{
		"licenseKey": licenseKey,
		"instanceId": instanceID,
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := c.client.Post(c.serverURL+"/api/v1/license/activate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("license server unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ValidateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &result, nil
}

func (c *Client) Heartbeat(licenseKey, instanceID string, tunnelCount, userCount int, cpuUsage, memUsage, diskUsage float64, ipAddress, domain string) error {
	reqBody := map[string]interface{}{
		"licenseKey":  licenseKey,
		"instanceId":  instanceID,
		"tunnelCount": tunnelCount,
		"userCount":   userCount,
		"cpuUsage":    cpuUsage,
		"memUsage":    memUsage,
		"diskUsage":   diskUsage,
		"ipAddress":   ipAddress,
		"domain":      domain,
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := c.client.Post(c.serverURL+"/api/v1/license/heartbeat", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("license server unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed: status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetLicenseInfo(licenseKey string) (*ValidateResponse, error) {
	resp, err := c.client.Get(c.serverURL + "/api/v1/license/info?licenseKey=" + licenseKey)
	if err != nil {
		return nil, fmt.Errorf("license server unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ValidateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &result, nil
}

func ParseExpiry(expiresAtStr string) *time.Time {
	if expiresAtStr == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil
	}
	return &t
}
