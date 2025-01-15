package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"goproxy/application"
	"log"
	"os"
	"time"
)

type RedisCache[T any] struct {
	client *redis.Client
}

func NewRedisCache[T any]() (application.CacheWithTTL[T], error) {
	host := os.Getenv("TC_CACHE_HOST")
	if host == "" {
		return nil, errors.New("env variable TC_CACHE_HOST is not set")
	}

	port := os.Getenv("TC_CACHE_PORT")
	if port == "" {
		return nil, errors.New("env variable TC_CACHE_PORT is not set")
	}

	user := os.Getenv("TC_CACHE_USER")
	if user == "" {
		return nil, errors.New("env variable TC_CACHE_USER is not set")
	}

	password := os.Getenv("TC_CACHE_PASSWORD")

	cacheClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Username: user,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	_, err := cacheClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	return &RedisCache[T]{
		client: cacheClient,
	}, nil
}

func (r *RedisCache[T]) Get(key string) (T, error) {
	var zero T
	ctx := context.Background()

	data, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return zero, fmt.Errorf("not found")
	}
	if err != nil {
		return zero, err
	}

	var value T
	deserializeErr := json.Unmarshal([]byte(data), &value)
	if deserializeErr != nil {
		return zero, deserializeErr
	}
	return value, nil
}

func (r *RedisCache[T]) Set(key string, value T) error {
	ctx := context.Background()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, 0).Err()
}

func (r *RedisCache[T]) Expire(key string, ttl time.Duration) error {
	ctx := context.Background()
	return r.client.Expire(ctx, key, ttl).Err()
}

func (r *RedisCache[T]) Delete(key string) error {
	_ = r.Expire(key, time.Second)
	return nil
}
