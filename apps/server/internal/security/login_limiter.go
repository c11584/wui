package security

import (
	"sync"
	"time"
)

type LoginAttempt struct {
	Count       int
	FirstTry    time.Time
	LastTry     time.Time
	LockedUntil time.Time
}

type LoginLimiter struct {
	attempts    map[string]*LoginAttempt
	maxAttempts int
	lockoutTime time.Duration
	windowTime  time.Duration
	mu          sync.RWMutex
}

func NewLoginLimiter(maxAttempts int, lockoutMinutes, windowMinutes int) *LoginLimiter {
	return &LoginLimiter{
		attempts:    make(map[string]*LoginAttempt),
		maxAttempts: maxAttempts,
		lockoutTime: time.Duration(lockoutMinutes) * time.Minute,
		windowTime:  time.Duration(windowMinutes) * time.Minute,
	}
}

func (l *LoginLimiter) RecordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	attempt, exists := l.attempts[ip]

	if !exists || now.Sub(attempt.FirstTry) > l.windowTime {
		attempt = &LoginAttempt{
			Count:    1,
			FirstTry: now,
			LastTry:  now,
		}
		l.attempts[ip] = attempt
	} else {
		attempt.Count++
		attempt.LastTry = now

		if attempt.Count >= l.maxAttempts {
			attempt.LockedUntil = now.Add(l.lockoutTime)
		}
	}
}

func (l *LoginLimiter) RecordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

func (l *LoginLimiter) IsLocked(ip string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		return false
	}

	if time.Now().Before(attempt.LockedUntil) {
		return true
	}

	return false
}

func (l *LoginLimiter) GetRemainingLockTime(ip string) time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		return 0
	}

	remaining := time.Until(attempt.LockedUntil)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (l *LoginLimiter) Cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for ip, attempt := range l.attempts {
		if now.Sub(attempt.LastTry) > l.lockoutTime*2 {
			delete(l.attempts, ip)
		}
	}
}

func (l *LoginLimiter) GetAttempts(ip string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if attempt, exists := l.attempts[ip]; exists {
		return attempt.Count
	}
	return 0
}

var GlobalLimiter = NewLoginLimiter(5, 15, 30)

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			GlobalLimiter.Cleanup()
		}
	}()
}
