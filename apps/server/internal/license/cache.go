package license

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type CacheInfo struct {
	LicenseKey  string `json:"license_key"`
	LastChecked int64  `json:"last_checked"`
}

type FileLicenseCache struct {
	FilePath    string
	GracePeriod time.Duration
	mu          sync.RWMutex
	Info        *CacheInfo
}

func NewFileLicenseCache(graceDays int, path string) *FileLicenseCache {
	c := &FileLicenseCache{FilePath: path, GracePeriod: time.Duration(graceDays) * 24 * time.Hour}
	c.load()
	return c
}

func (c *FileLicenseCache) load() {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, err := os.Open(c.FilePath)
	if err != nil {
		return
	}
	defer f.Close()
	var info CacheInfo
	if err := json.NewDecoder(f).Decode(&info); err == nil {
		c.Info = &info
	}
}

func (c *FileLicenseCache) save() {
	f, err := os.Create(c.FilePath)
	if err != nil {
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(c.Info)
}

func (c *FileLicenseCache) Save(licenseKey string, at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Info = &CacheInfo{LicenseKey: licenseKey, LastChecked: at.Unix()}
	c.save()
}

func (c *FileLicenseCache) GetWithinGrace(now time.Time) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Info == nil {
		return "", false
	}
	last := time.Unix(c.Info.LastChecked, 0)
	if now.Sub(last) <= c.GracePeriod {
		return c.Info.LicenseKey, true
	}
	return "", false
}

func (c *FileLicenseCache) UpdateHeartbeat(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Info != nil {
		c.Info.LastChecked = now.Unix()
		c.save()
	}
}
