package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"goproxy/application"
	"goproxy/domain/events"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"time"
)

type TrafficCollector struct {
	cache      *redis.Client
	messageBus application.MessageBusService
}

func NewTrafficCollector() (*TrafficCollector, error) {
	messageBus, err := instantiateMessageBusService()
	if err != nil {
		return nil, err
	}

	redisClient, err := instantiateCache()
	if err != nil {
		_ = messageBus.Close()
		return nil, err
	}

	return &TrafficCollector{
		cache:      redisClient,
		messageBus: messageBus,
	}, nil
}

func instantiateMessageBusService() (application.MessageBusService, error) {
	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	groupId := os.Getenv("KAFKA_GROUP_ID")
	autoOffsetReset := os.Getenv("KAFKA_AUTO_OFFSET_RESET")
	topic := os.Getenv("KAFKA_TOPIC")

	if groupId == "" || autoOffsetReset == "" || topic == "" || bootstrapServers == "" {
		return nil, fmt.Errorf("invalid configuration")
	}

	messageBusService, err := services.NewKafkaService(bootstrapServers, groupId, autoOffsetReset)
	if err != nil {
		return nil, err
	}

	return messageBusService, nil
}

func instantiateCache() (*redis.Client, error) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		return nil, errors.New("env variable REDIS_HOST is not set")
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		return nil, errors.New("env variable REDIS_PORT is not set")
	}

	redisUser := os.Getenv("REDIS_USER")
	if redisUser == "" {
		return nil, errors.New("env variable REDIS_USER is not set")
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Username: redisUser,
		Password: redisPassword,
		DB:       0,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	log.Println("Connected to Redis successfully")
	return redisClient, nil
}

func (t *TrafficCollector) ProcessEvents() {
	defer func(messageBus application.MessageBusService) {
		_ = messageBus.Close()
	}(t.messageBus)

	err := t.messageBus.Subscribe([]string{os.Getenv("KAFKA_TOPIC")})
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %s", err)
	}

	for {
		event, readErr := t.messageBus.Consume()
		if readErr == nil {
			t.consume(event)
		} else {
			log.Printf("Consumer error: %v (%v)\n", readErr, event)
		}
	}
}

func (t *TrafficCollector) consume(outboxEvent *events.OutboxEvent) {
	var event events.UserConsumedTrafficEvent
	err := json.Unmarshal([]byte(outboxEvent.Payload), &event)
	if err != nil {
		log.Printf("Invalid event: %v", err)
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf("user:%d:traffic:%s", event.UserId, time.Now().Format("02-01-2006"))

	currentTraffic, err := t.cache.HGetAll(ctx, key).Result()
	if err != nil {
		log.Printf("failed to get current traffic: %v", err)
		return
	}

	var inBytes, outBytes int
	if len(currentTraffic) > 0 {
		_, currentInScanErr := fmt.Sscanf(currentTraffic["inBytes"], "%d", &inBytes)
		if currentInScanErr != nil {
			inBytes = 0
		}
		_, currentOutScanErr := fmt.Sscanf(currentTraffic["outBytes"], "%d", &outBytes)
		if currentOutScanErr != nil {
			outBytes = 0
		}
	}

	inBytes += event.InBytes
	outBytes += event.OutBytes

	err = t.cache.HSet(ctx, key, map[string]interface{}{
		"inBytes":  inBytes,
		"outBytes": outBytes,
	}).Err()
	if err != nil {
		log.Printf("failed to update traffic: %v", err)
		return
	}

	err = t.cache.Expire(ctx, key, 24*time.Hour).Err()
	if err != nil {
		log.Printf("failed to set TTL: %v", err)
		return
	}
}
