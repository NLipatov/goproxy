package services

import (
	"fmt"
	"github.com/cespare/xxhash/v2"
	serviceconfigurations "goproxy/infrastructure/config"
	"sync"
	"sync/atomic"
	"time"
)

type request struct {
	userID int
	target string
	tokens int64
	resp   chan bool
}

type doneRequest struct {
	userID int
	target string
}

// RateLimiter implements a sharded rate limiter.
type RateLimiter struct {
	reqChan    chan request
	doneChan   chan doneRequest
	stopChan   chan struct{}
	shards     []*shard
	maxConns   int
	blockDur   time.Duration
	cleanupInt time.Duration

	shardsCapacity int64   // Capacity per bucket
	shardsFillRate float64 // Fill rate per bucket

	wg sync.WaitGroup
}

type shard struct {
	buckets map[string]*tokenBucket
	mu      sync.Mutex
}

type tokenBucket struct {
	capacity       int64
	fillRate       float64
	available      int64
	lastUpdateNano int64
	connections    int64
	blockUntil     int64
}

// NewRateLimiter initializes a new RateLimiter with the given configuration.
func NewRateLimiter(config serviceconfigurations.RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		reqChan:        make(chan request, 1024),
		doneChan:       make(chan doneRequest, 1024),
		stopChan:       make(chan struct{}),
		maxConns:       config.MaxConns,
		blockDur:       config.BlockDur,
		cleanupInt:     config.CleanupInt,
		shardsCapacity: config.Capacity,
		shardsFillRate: config.FillRate,
		shards:         make([]*shard, config.ShardCount),
	}

	for i := 0; i < config.ShardCount; i++ {
		rl.shards[i] = &shard{buckets: make(map[string]*tokenBucket)}
	}

	rl.wg.Add(1)
	go rl.loop()

	return rl
}

// Stop gracefully shuts down the RateLimiter.
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
	rl.wg.Wait()
}

// Allow checks if a user can perform an action with the specified token cost.
func (rl *RateLimiter) Allow(userID int, target string, tokens int64) bool {
	respChan := make(chan bool, 1)
	req := request{
		userID: userID,
		target: target,
		tokens: tokens,
		resp:   respChan,
	}
	rl.reqChan <- req
	return <-respChan
}

// Done signals that a user has completed an action.
func (rl *RateLimiter) Done(userID int, target string) {
	rl.doneChan <- doneRequest{userID, target}
}

func (rl *RateLimiter) loop() {
	defer rl.wg.Done()
	ticker := time.NewTicker(rl.cleanupInt)
	defer ticker.Stop()

	for {
		select {
		case req := <-rl.reqChan:
			allowed := rl.allowInternal(req.userID, req.target, req.tokens)
			req.resp <- allowed

		case dreq := <-rl.doneChan:
			rl.doneInternal(dreq.userID, dreq.target)

		case <-ticker.C:
			rl.cleanupInternal()

		case <-rl.stopChan:
			return
		}
	}
}

func (rl *RateLimiter) allowInternal(userID int, target string, tokens int64) bool {
	// Create a unique key for user and target
	key := fmt.Sprintf("%d|%s", userID, target)
	s := rl.getShard(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, ok := s.buckets[key]
	if !ok {
		// Create a new token bucket with capacity and fill rate from config
		bucket = &tokenBucket{
			capacity:       rl.shardsCapacity, // e.g., 100 MB
			fillRate:       rl.shardsFillRate, // e.g., 125 MB/s
			available:      rl.shardsCapacity,
			lastUpdateNano: time.Now().UnixNano(),
		}
		s.buckets[key] = bucket
	}

	now := time.Now().UnixNano()
	// Check if the bucket is blocked
	if now < atomic.LoadInt64(&bucket.blockUntil) {
		return false
	}
	// Refill tokens
	bucket.refill(now)

	// Check for maximum concurrent connections
	if bucket.connections >= int64(rl.maxConns) {
		bucket.blockUntil = now + int64(rl.blockDur)
		return false
	}

	// Check if there are enough tokens
	if bucket.available >= tokens {
		bucket.available -= tokens
		bucket.connections++
		return true
	}

	// If not enough tokens, block the user
	bucket.blockUntil = now + int64(rl.blockDur)
	return false
}

func (rl *RateLimiter) doneInternal(userID int, target string) {
	key := fmt.Sprintf("%d|%s", userID, target)
	s := rl.getShard(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	if bucket, ok := s.buckets[key]; ok && bucket.connections > 0 {
		bucket.connections--
	}
}

func (rl *RateLimiter) cleanupInternal() {
	now := time.Now().UnixNano()
	for _, s := range rl.shards {
		s.mu.Lock()
		for key, bucket := range s.buckets {
			// Unblock if block duration has passed
			if bucket.blockUntil > 0 && now > bucket.blockUntil {
				bucket.blockUntil = 0
			}
			// Remove bucket if no active connections and it hasn't been updated for 5 minutes
			if bucket.connections == 0 && (now-bucket.lastUpdateNano) > int64(5*time.Minute) {
				delete(s.buckets, key)
			}
		}
		s.mu.Unlock()
	}
}

func (rl *RateLimiter) getShard(key string) *shard {
	idx := xxhash32(key) % uint32(len(rl.shards))
	return rl.shards[idx]
}

func (tb *tokenBucket) refill(nowNano int64) {
	elapsedSec := float64(nowNano-tb.lastUpdateNano) / float64(time.Second)
	if elapsedSec <= 0 {
		return
	}
	tb.lastUpdateNano = nowNano

	newTokens := int64(elapsedSec * tb.fillRate)
	if newTokens > 0 {
		tb.available += newTokens
		if tb.available > tb.capacity {
			tb.available = tb.capacity
		}
	}
}

func xxhash32(key string) uint32 {
	return uint32(xxhash.Sum64([]byte(key)))
}
