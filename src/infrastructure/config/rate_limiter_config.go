package config

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

// RateLimiterConfig holds the configuration settings for the application.
type RateLimiterConfig struct {
	NetBandwidthMbps int64         // Network bandwidth in megabits per second (e.g., 1000 for 1 Gigabit/s)
	MaxUsers         int           // Maximum number of users
	ShardCount       int           // Number of shards for rate limiting
	Capacity         int64         // TokenBucket capacity in bytes
	FillRate         float64       // Token fill rate in bytes per second
	MaxConns         int           // Maximum number of concurrent connections
	BlockDur         time.Duration // Duration to block a user when rate limit is exceeded
	CleanupInt       time.Duration // Interval to clean up old buckets
}

// LoadRateLimiterConfig reads configuration from environment variables and applies default values if necessary.
// It expects:
// - NET_BANDWIDTH in megabits per second (e.g., "1000" for 1 Gigabit/s)
// - MAX_USERS as an integer (e.g., "100")
// - SHARD_COUNT as an integer (optional; defaults based on CPU cores)
// - Other settings can also be configured via environment variables.
func LoadRateLimiterConfig() RateLimiterConfig {
	var config RateLimiterConfig
	var err error

	// Load NET_BANDWIDTH
	netBandwidthStr := os.Getenv("NET_BANDWIDTH")
	if netBandwidthStr == "" {
		log.Println("Environment variable NET_BANDWIDTH not set. Using default value: 1000 Mbps (1 Gigabit/s)")
		config.NetBandwidthMbps = 1000 // Default to 1 Gigabit/s
	} else {
		config.NetBandwidthMbps, err = strconv.ParseInt(netBandwidthStr, 10, 64)
		if err != nil || config.NetBandwidthMbps <= 0 {
			log.Printf("Invalid NET_BANDWIDTH value: %s. Using default value: 1000 Mbps (1 Gigabit/s)\n", netBandwidthStr)
			config.NetBandwidthMbps = 1000
		}
	}

	// Load MAX_USERS
	maxUsersStr := os.Getenv("MAX_USERS")
	if maxUsersStr == "" {
		log.Println("Environment variable MAX_USERS not set. Using default value: 100 users")
		config.MaxUsers = 100
	} else {
		config.MaxUsers, err = strconv.Atoi(maxUsersStr)
		if err != nil || config.MaxUsers <= 0 {
			log.Printf("Invalid MAX_USERS value: %s. Using default value: 100 users\n", maxUsersStr)
			config.MaxUsers = 100
		}
	}

	// Load SHARD_COUNT
	shardCountStr := os.Getenv("SHARD_COUNT")
	if shardCountStr == "" {
		defaultShardCount := runtime.NumCPU() * 2 // 2 shards per CPU core
		log.Printf("Environment variable SHARD_COUNT not set. Using default value based on CPU cores: %d shards\n", defaultShardCount)
		config.ShardCount = defaultShardCount
	} else {
		config.ShardCount, err = strconv.Atoi(shardCountStr)
		if err != nil || config.ShardCount <= 0 {
			defaultShardCount := runtime.NumCPU() * 2
			log.Printf("Invalid SHARD_COUNT value: %s. Using default value based on CPU cores: %d shards\n", shardCountStr, defaultShardCount)
			config.ShardCount = defaultShardCount
		}
	}

	// Set capacity and fillRate based on NET_BANDWIDTH
	// Convert megabits per second (Mbps) to bytes per second (Bps)
	// 1 Mbps = 125,000 bytes per second
	config.Capacity = 100 * 1024 * 1024                         // 100 MB by default
	config.FillRate = float64(config.NetBandwidthMbps) * 125000 // Convert Mbps to bytes per second

	// Cap fillRate to a maximum value (e.g., 125 MB/s for 1 Gigabit/s)
	maxFillRate := 125 * 1024 * 1024 // 125 MB/s
	if config.FillRate > float64(maxFillRate) {
		log.Printf("FillRate exceeds maximum value (%d bytes/sec). Setting FillRate to %d bytes/sec\n", maxFillRate, maxFillRate)
		config.FillRate = float64(maxFillRate)
	}

	// Optionally, allow overriding Capacity and FillRate via environment variables
	// Load CAPACITY
	capacityStr := os.Getenv("CAPACITY_MB")
	if capacityStr != "" {
		capacityMB, err := strconv.ParseInt(capacityStr, 10, 64)
		if err == nil && capacityMB > 0 {
			config.Capacity = capacityMB * 1024 * 1024 // Convert MB to bytes
			log.Printf("Overriding Capacity with CAPACITY_MB: %d MB\n", capacityMB)
		} else {
			log.Printf("Invalid CAPACITY_MB value: %s. Using existing Capacity: %d bytes\n", capacityStr, config.Capacity)
		}
	}

	// Load FILL_RATE
	fillRateStr := os.Getenv("FILL_RATE_MBPS")
	if fillRateStr != "" {
		fillRateMbps, err := strconv.ParseFloat(fillRateStr, 64)
		if err == nil && fillRateMbps > 0 {
			config.FillRate = fillRateMbps * 125000 // Convert Mbps to bytes per second
			log.Printf("Overriding FillRate with FILL_RATE_MBPS: %.2f Mbps (%.0f bytes/sec)\n", fillRateMbps, config.FillRate)
		} else {
			log.Printf("Invalid FILL_RATE_MBPS value: %s. Using existing FillRate: %.0f bytes/sec\n", fillRateStr, config.FillRate)
		}
	}

	// Set MaxConns
	config.MaxConns = config.MaxUsers * 25 // For example, 25 connections per user

	// Optionally, allow overriding MaxConns via environment variables
	maxConnsStr := os.Getenv("MAX_CONNS")
	if maxConnsStr != "" {
		maxConns, err := strconv.Atoi(maxConnsStr)
		if err == nil && maxConns > 0 {
			config.MaxConns = maxConns
			log.Printf("Overriding MaxConns with MAX_CONNS: %d\n", maxConns)
		} else {
			log.Printf("Invalid MAX_CONNS value: %s. Using existing MaxConns: %d\n", maxConnsStr, config.MaxConns)
		}
	}

	// Load BLOCK_DURATION
	blockDurStr := os.Getenv("BLOCK_DURATION_SEC")
	if blockDurStr == "" {
		config.BlockDur = 30 * time.Second // Default
	} else {
		blockDurSec, err := strconv.Atoi(blockDurStr)
		if err == nil && blockDurSec > 0 {
			config.BlockDur = time.Duration(blockDurSec) * time.Second
			log.Printf("Overriding BlockDur with BLOCK_DURATION_SEC: %d seconds\n", blockDurSec)
		} else {
			log.Printf("Invalid BLOCK_DURATION_SEC value: %s. Using default BlockDur: %v\n", blockDurStr, config.BlockDur)
		}
	}

	// Load CLEANUP_INTERVAL
	cleanupIntStr := os.Getenv("CLEANUP_INTERVAL_SEC")
	if cleanupIntStr == "" {
		config.CleanupInt = 1 * time.Minute
	} else {
		cleanupIntSec, err := strconv.Atoi(cleanupIntStr)
		if err == nil && cleanupIntSec > 0 {
			config.CleanupInt = time.Duration(cleanupIntSec) * time.Second
			log.Printf("Overriding CleanupInt with CLEANUP_INTERVAL_SEC: %d seconds\n", cleanupIntSec)
		} else {
			log.Printf("Invalid CLEANUP_INTERVAL_SEC value: %s. Using default CleanupInt: %v\n", cleanupIntStr, config.CleanupInt)
		}
	}

	return config
}
