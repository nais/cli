// Package mcp provides the MCP server implementation for Nais CLI.
package mcp

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter.
// requestsPerMinute specifies the maximum requests per minute.
// If requestsPerMinute is 0 or negative, the limiter allows all requests.
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	if requestsPerMinute <= 0 {
		return &RateLimiter{
			maxTokens:  -1, // unlimited
			refillRate: 0,
		}
	}

	maxTokens := float64(requestsPerMinute)
	refillRate := float64(requestsPerMinute) / 60.0 // tokens per second

	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit.
// Returns true if the request is allowed, false if rate limited.
func (r *RateLimiter) Allow() bool {
	// Unlimited rate limiter
	if r.maxTokens < 0 {
		return true
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens >= 1 {
		r.tokens--
		return true
	}

	return false
}

// refill adds tokens based on elapsed time since last refill.
// Must be called with mu held.
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.lastRefill = now

	r.tokens += elapsed * r.refillRate
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}
}

// TokensAvailable returns the current number of tokens available.
// Useful for debugging and monitoring.
func (r *RateLimiter) TokensAvailable() float64 {
	if r.maxTokens < 0 {
		return -1 // unlimited
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()
	return r.tokens
}

// WaitTime returns how long to wait before a request would be allowed.
// Returns 0 if a request is currently allowed or if unlimited.
func (r *RateLimiter) WaitTime() time.Duration {
	if r.maxTokens < 0 {
		return 0
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens >= 1 {
		return 0
	}

	// Calculate time needed to get 1 token
	tokensNeeded := 1 - r.tokens
	secondsNeeded := tokensNeeded / r.refillRate

	return time.Duration(secondsNeeded * float64(time.Second))
}
