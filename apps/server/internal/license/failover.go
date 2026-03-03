package license

import (
	"sync"
	"time"
)

type LicenseCache struct {
	Key         string
	InstanceID  string
	IsValid     bool
	Type        string
	Plan        string
	MaxTunnels  int
	MaxUsers    int
	MaxTraffic  int64
	Features    string
	ExpiresAt   *time.Time
	LastChecked time.Time
	ActivatedAt time.Time
}

type CacheManager struct {
	mu          sync.RWMutex
	caches      map[uint]*LicenseCache
	gracePeriod time.Duration
}

func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches:      make(map[uint]*LicenseCache),
		gracePeriod: 72 * time.Hour,
	}
}

func (cm *CacheManager) Set(userID uint, cache *LicenseCache) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cache.LastChecked = time.Now()
	cm.caches[userID] = cache
}

func (cm *CacheManager) Get(userID uint) (*LicenseCache, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cache, exists := cm.caches[userID]
	return cache, exists
}

func (cm *CacheManager) IsValidWithGracePeriod(userID uint) (bool, *LicenseCache) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cache, exists := cm.caches[userID]
	if !exists {
		return false, nil
	}

	if cache.IsValid {
		if cache.ExpiresAt != nil && time.Now().After(*cache.ExpiresAt) {
			return false, cache
		}
		return true, cache
	}

	if time.Since(cache.LastChecked) < cm.gracePeriod {
		return true, cache
	}

	return false, cache
}

func (cm *CacheManager) Remove(userID uint) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.caches, userID)
}

func (cm *CacheManager) SetGracePeriod(duration time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.gracePeriod = duration
}

func (cm *CacheManager) GetGracePeriod() time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.gracePeriod
}

var globalCacheManager = NewCacheManager()

func GetCacheManager() *CacheManager {
	return globalCacheManager
}
