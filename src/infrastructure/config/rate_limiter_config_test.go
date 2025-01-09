package config

import (
	"os"
	"runtime"
	"testing"
	"time"
)

func TestLoadRateLimiterConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig RateLimiterConfig
	}{
		{
			name:    "Default Config",
			envVars: map[string]string{},
			expectedConfig: RateLimiterConfig{
				NetBandwidthMbps: 1000,
				MaxUsers:         100,
				ShardCount:       runtime.NumCPU() * 2,
				Capacity:         100 * 1024 * 1024,
				FillRate:         1000 * 125000,
				MaxConns:         100 * 10,
				BlockDur:         30 * time.Second,
				CleanupInt:       1 * time.Minute,
			},
		},
		{
			name: "Custom Config",
			envVars: map[string]string{
				"NET_BANDWIDTH":        "2000",
				"MAX_USERS":            "50",
				"SHARD_COUNT":          "8",
				"CAPACITY_MB":          "200",
				"FILL_RATE_MBPS":       "1500",
				"MAX_CONNS":            "500",
				"BLOCK_DURATION_SEC":   "60",
				"CLEANUP_INTERVAL_SEC": "120",
			},
			expectedConfig: RateLimiterConfig{
				NetBandwidthMbps: 2000,
				MaxUsers:         50,
				ShardCount:       8,
				Capacity:         200 * 1024 * 1024,
				FillRate:         1500 * 125000,
				MaxConns:         500,
				BlockDur:         60 * time.Second,
				CleanupInt:       2 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Load configuration
			config := LoadRateLimiterConfig()

			// Validate configuration
			if config != tt.expectedConfig {
				t.Errorf("expected %+v, got %+v", tt.expectedConfig, config)
			}

			// Clear environment variables
			for key := range tt.envVars {
				_ = os.Unsetenv(key)
			}
		})
	}
}
