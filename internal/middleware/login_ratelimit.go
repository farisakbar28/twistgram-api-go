package middleware

import (
	"sync"
	"time"
)

type loginAttempt struct {
	fails         int
	lockoutUntil time.Time
}

var (
	loginAttempts = make(map[string]*loginAttempt)
	loginMu       sync.Mutex
)

const (
	maxLoginFails    = 5
	loginCooldown    = 15 * time.Minute
)

// IsLoginLocked checks if an identifier is currently locked out.
func IsLoginLocked(identifier string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()
	la, exists := loginAttempts[identifier]
	if !exists {
		return false
	}
	if la.fails >= maxLoginFails && time.Now().Before(la.lockoutUntil) {
		return true
	}
	return false
}

// RecordLoginFailure increments the fail counter for an identifier.
func RecordLoginFailure(identifier string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	la, exists := loginAttempts[identifier]
	if !exists {
		la = &loginAttempt{}
		loginAttempts[identifier] = la
	}
	la.fails++
	if la.fails >= maxLoginFails {
		la.lockoutUntil = time.Now().Add(loginCooldown)
	}
}

// ResetLoginAttempts clears the fail counter after successful login.
func ResetLoginAttempts(identifier string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	delete(loginAttempts, identifier)
}
