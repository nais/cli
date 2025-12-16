package mcp

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Unlimited(t *testing.T) {
	// A rate limit of 0 should allow unlimited requests
	rl := NewRateLimiter(0)

	for i := 0; i < 1000; i++ {
		if !rl.Allow() {
			t.Errorf("unlimited rate limiter should allow all requests, blocked at request %d", i)
		}
	}

	// Tokens should return -1 for unlimited
	if tokens := rl.TokensAvailable(); tokens != -1 {
		t.Errorf("expected -1 tokens for unlimited, got %f", tokens)
	}

	// Wait time should be 0 for unlimited
	if wait := rl.WaitTime(); wait != 0 {
		t.Errorf("expected 0 wait time for unlimited, got %v", wait)
	}
}

func TestRateLimiter_NegativeLimit(t *testing.T) {
	// A negative rate limit should also be treated as unlimited
	rl := NewRateLimiter(-5)

	for i := 0; i < 100; i++ {
		if !rl.Allow() {
			t.Errorf("negative rate limit should allow all requests, blocked at request %d", i)
		}
	}
}

func TestRateLimiter_BasicLimit(t *testing.T) {
	// 60 requests per minute = 1 request per second
	rl := NewRateLimiter(60)

	// Should have 60 tokens initially
	initialTokens := rl.TokensAvailable()
	if initialTokens < 59 || initialTokens > 60 {
		t.Errorf("expected ~60 initial tokens, got %f", initialTokens)
	}

	// Use all tokens
	allowed := 0
	for i := 0; i < 100; i++ {
		if rl.Allow() {
			allowed++
		}
	}

	// Should have allowed approximately 60 requests (the initial bucket)
	if allowed < 59 || allowed > 61 {
		t.Errorf("expected ~60 allowed requests, got %d", allowed)
	}

	// Next request should be blocked
	if rl.Allow() {
		t.Error("expected request to be blocked after exhausting tokens")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	// 60 requests per minute = 1 token per second
	rl := NewRateLimiter(60)

	// Exhaust all tokens
	for rl.Allow() {
		// drain
	}

	// Verify we're blocked
	if rl.Allow() {
		t.Error("expected to be blocked after exhausting tokens")
	}

	// Wait a bit for refill (1 second should give us ~1 token)
	time.Sleep(1100 * time.Millisecond)

	// Should now be allowed
	if !rl.Allow() {
		t.Error("expected request to be allowed after waiting for refill")
	}
}

func TestRateLimiter_WaitTime(t *testing.T) {
	// 60 requests per minute = 1 token per second
	rl := NewRateLimiter(60)

	// Exhaust all tokens
	for rl.Allow() {
		// drain
	}

	// Wait time should be positive but less than 2 seconds (need 1 token at 1/sec)
	wait := rl.WaitTime()
	if wait <= 0 {
		t.Errorf("expected positive wait time, got %v", wait)
	}
	if wait > 2*time.Second {
		t.Errorf("expected wait time less than 2 seconds, got %v", wait)
	}
}

func TestRateLimiter_WaitTime_WhenAvailable(t *testing.T) {
	rl := NewRateLimiter(60)

	// Should have tokens available, so wait time should be 0
	if wait := rl.WaitTime(); wait != 0 {
		t.Errorf("expected 0 wait time when tokens available, got %v", wait)
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	// Test concurrent access to the rate limiter
	rl := NewRateLimiter(100)

	var wg sync.WaitGroup
	allowed := make(chan bool, 200)

	// Spawn 200 goroutines, each trying to make a request
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- rl.Allow()
		}()
	}

	wg.Wait()
	close(allowed)

	// Count allowed requests
	allowedCount := 0
	for a := range allowed {
		if a {
			allowedCount++
		}
	}

	// Should have allowed approximately 100 requests (the bucket size)
	if allowedCount < 95 || allowedCount > 105 {
		t.Errorf("expected ~100 allowed requests in concurrent scenario, got %d", allowedCount)
	}
}

func TestRateLimiter_TokensAvailable(t *testing.T) {
	rl := NewRateLimiter(10)

	// Initially should have ~10 tokens
	tokens := rl.TokensAvailable()
	if tokens < 9 || tokens > 10 {
		t.Errorf("expected ~10 tokens initially, got %f", tokens)
	}

	// Use 3 tokens
	rl.Allow()
	rl.Allow()
	rl.Allow()

	tokens = rl.TokensAvailable()
	if tokens < 6 || tokens > 8 {
		t.Errorf("expected ~7 tokens after using 3, got %f", tokens)
	}
}

func TestRateLimiter_DoesNotExceedMax(t *testing.T) {
	rl := NewRateLimiter(10)

	// Wait a long time (tokens should not exceed max)
	time.Sleep(200 * time.Millisecond)

	tokens := rl.TokensAvailable()
	if tokens > 10 {
		t.Errorf("tokens should not exceed max of 10, got %f", tokens)
	}
}
