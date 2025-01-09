package services

import (
	"goproxy/infrastructure/config"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rlConf := config.RateLimiterConfig{
		MaxConns:   10,
		BlockDur:   1 * time.Second,
		CleanupInt: 10 * time.Second,
		Capacity:   100,
		FillRate:   50,
		ShardCount: 2,
	}

	rl := NewRateLimiter(rlConf)
	defer rl.Stop()

	userID := 1
	target := "test_target"
	tokens := int64(30)

	// Allow the first request
	if !rl.Allow(userID, target, tokens) {
		t.Fatalf("expected Allow to return true for the first request")
	}

	// Exceed the bucket capacity
	if rl.Allow(userID, target, 80) {
		t.Fatalf("expected Allow to return false when exceeding bucket capacity")
	}

	// Wait for refill and allow again
	time.Sleep(3 * time.Second)
	if !rl.Allow(userID, target, tokens) {
		t.Fatalf("expected Allow to return true after refill")
	}
}

func TestRateLimiter_Done(t *testing.T) {
	rlConf := config.RateLimiterConfig{
		MaxConns:   1,
		BlockDur:   100 * time.Millisecond,
		CleanupInt: 100 * time.Millisecond,
		Capacity:   100,
		FillRate:   50,
		ShardCount: 2,
	}

	rl := NewRateLimiter(rlConf)
	defer rl.Stop()

	userID := 1
	target := "test_target"
	tokens := int64(30)

	if !rl.Allow(userID, target, tokens) {
		t.Fatalf("expected Allow to return true for the first request")
	}

	// Exceed connection limit
	if rl.Allow(userID, target, tokens) {
		t.Fatalf("expected Allow to return false due to max connections")
	}

	// Signal Done and allow again
	rl.Done(userID, target)
	time.Sleep(rlConf.BlockDur * 2)
	if !rl.Allow(userID, target, tokens) {
		t.Fatalf("expected Allow to return true after Done")
	}
}

func TestRateLimiter_Block(t *testing.T) {
	rlConf := config.RateLimiterConfig{
		MaxConns:   10,
		BlockDur:   100 * time.Millisecond,
		CleanupInt: 100 * time.Millisecond,
		Capacity:   100,
		FillRate:   50,
		ShardCount: 2,
	}

	rl := NewRateLimiter(rlConf)
	defer rl.Stop()

	userID := 1
	target := "test_target"
	tokens := int64(150) // Exceeds capacity

	if rl.Allow(userID, target, tokens) {
		t.Fatalf("expected Allow to return false due to block")
	}

	// Ensure blocked requests are denied
	if rl.Allow(userID, target, 10) {
		t.Fatalf("expected Allow to return false for blocked user")
	}

	// Wait for block duration and allow
	time.Sleep(rlConf.CleanupInt * 2)
	if !rl.Allow(userID, target, 10) {
		t.Fatalf("expected Allow to return true after block duration")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rlConf := config.RateLimiterConfig{
		MaxConns:   10,
		BlockDur:   100 * time.Millisecond,
		CleanupInt: 100 * time.Millisecond,
		Capacity:   100,
		FillRate:   50,
		ShardCount: 2,
	}

	rl := NewRateLimiter(rlConf)
	defer rl.Stop()

	userID := 1
	target := "test_target"
	tokens := int64(10)

	if !rl.Allow(userID, target, tokens) {
		t.Fatalf("expected Allow to return true for the first request")
	}

	time.Sleep(rlConf.CleanupInt * 2) // Wait for cleanup interval

	// Ensure the bucket is cleaned up
	rl.doneInternal(userID, target) // Simulate request completion
}
