package security

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type Captcha struct {
	Code      string
	ExpiresAt time.Time
}

type CaptchaManager struct {
	captchas map[string]*Captcha
	ttl      time.Duration
	mu       sync.RWMutex
}

func NewCaptchaManager(ttlSeconds int) *CaptchaManager {
	m := &CaptchaManager{
		captchas: make(map[string]*Captcha),
		ttl:      time.Duration(ttlSeconds) * time.Second,
	}

	go m.cleanup()
	return m
}

func (m *CaptchaManager) Generate() (id string, code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idBytes := make([]byte, 16)
	rand.Read(idBytes)
	id = base64.URLEncoding.EncodeToString(idBytes)

	codeBytes := make([]byte, 3)
	rand.Read(codeBytes)
	code = base64.URLEncoding.EncodeToString(codeBytes)[:6]

	for i := range code {
		if code[i] >= 'a' && code[i] <= 'z' {
			code = code[:i] + string(code[i]-'a'+'A') + code[i+1:]
		}
	}

	m.captchas[id] = &Captcha{
		Code:      code,
		ExpiresAt: time.Now().Add(m.ttl),
	}

	return id, code
}

func (m *CaptchaManager) Verify(id string, code string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	captcha, exists := m.captchas[id]
	if !exists {
		return false
	}

	delete(m.captchas, id)

	if time.Now().After(captcha.ExpiresAt) {
		return false
	}

	return captcha.Code == code
}

func (m *CaptchaManager) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, captcha := range m.captchas {
			if now.After(captcha.ExpiresAt) {
				delete(m.captchas, id)
			}
		}
		m.mu.Unlock()
	}
}

var GlobalCaptcha = NewCaptchaManager(300)
