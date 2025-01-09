package services

import (
	"testing"
	"time"
)

func TestMapCacheWithTTL_SetAndGet(t *testing.T) {
	cache := NewMapCacheWithTTL[string]()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}

	val, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error on Get: %v", err)
	}
	if val != "value1" {
		t.Fatalf("expected value 'value1', got '%s'", val)
	}
}

func TestMapCacheWithTTL_Expire(t *testing.T) {
	cache := NewMapCacheWithTTL[string]()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}

	err = cache.Expire("key1", time.Millisecond*100)
	if err != nil {
		t.Fatalf("unexpected error on Expire: %v", err)
	}

	val, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error on Get before expiration: %v", err)
	}
	if val != "value1" {
		t.Fatalf("expected value 'value1', got '%s'", val)
	}

	time.Sleep(time.Millisecond * 100 * 3)

	val, err = cache.Get("key1")
	if err == nil {
		t.Fatalf("expected error on Get after expiration")
	}

	if val != "" {
		t.Fatalf("expected empty value, got '%s'", val)
	}
}

func TestMapCacheWithTTL_GetNonExistentKey(t *testing.T) {
	cache := NewMapCacheWithTTL[string]()

	val, err := cache.Get("key1")
	if err == nil {
		t.Fatalf("expected error on Get before Set")
	}

	if err.Error() != "key not found" {
		t.Fatalf("expected error 'key not found', got: %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty value, got '%s'", val)
	}
}

func TestMapCacheWithTTL_ExpireNonExistentKey(t *testing.T) {
	cache := NewMapCacheWithTTL[string]()

	err := cache.Expire("key1", 1*time.Second)
	if err == nil {
		t.Fatalf("expected error on Expire before Set")
	}

	if err.Error() != "key not found" {
		t.Fatalf("expected error 'key not found', got: %v", err)
	}
}

func TestMapCacheWithTTL_UpdateTTL(t *testing.T) {
	cache := NewMapCacheWithTTL[string]()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}

	err = cache.Expire("key1", 1*time.Second)
	if err != nil {
		t.Fatalf("unexpected error on Expire: %v", err)
	}

	err = cache.Expire("key1", 3*time.Second)
	if err != nil {
		t.Fatalf("unexpected error on Expire: %v", err)
	}

	time.Sleep(2 * time.Second)
	val, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error on Get after TTL update: %v", err)
	}
	if val != "value1" {
		t.Fatalf("expected value 'value1', got '%s'", val)
	}

	time.Sleep(2 * time.Second)
	val, err = cache.Get("key1")
	if err == nil {
		t.Fatalf("expected error on Get after expiration")
	}

	if err.Error() != "key not found (expired)" {
		t.Fatalf("expected error 'key not found (expired)', got: %v", err)
	}

	if val != "" {
		t.Fatalf("expected empty value, got '%s'", val)
	}
}
