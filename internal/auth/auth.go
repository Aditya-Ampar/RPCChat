package auth

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type Token struct {
	Value     string
	Username  string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type Authenticator struct {
	mu            sync.RWMutex
	activeTokens  map[string]*Token
	tokenSecret   string
	tokenDuration time.Duration
}

func NewAuthenticator(secret string) *Authenticator {
	return &Authenticator{
		activeTokens:  make(map[string]*Token),
		tokenSecret:   secret,
		tokenDuration: 1 * time.Hour,
	}
}

func (a *Authenticator) IssueToken(username string) (*Token, error) {
	if username == "" {
		return nil, fmt.Errorf("username required")
	}

	now := time.Now()
	tokenStr := fmt.Sprintf("%x", sha256.Sum256([]byte(username+a.tokenSecret+now.String())))

	token := &Token{
		Value:     tokenStr,
		Username:  username,
		IssuedAt:  now,
		ExpiresAt: now.Add(a.tokenDuration),
	}

	a.mu.Lock()
	a.activeTokens[tokenStr] = token
	a.mu.Unlock()

	return token, nil
}

func (a *Authenticator) Validate(tokenStr string) (string, error) {
	a.mu.RLock()
	token, exists := a.activeTokens[tokenStr]
	a.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("invalid token")
	}

	if time.Now().After(token.ExpiresAt) {
		a.mu.Lock()
		delete(a.activeTokens, tokenStr)
		a.mu.Unlock()
		return "", fmt.Errorf("token expired")
	}
	return token.Username, nil
}

func (a *Authenticator) Revoke(tokenStr string) {
	a.mu.Lock()
	delete(a.activeTokens, tokenStr)
	a.mu.Unlock()
}
