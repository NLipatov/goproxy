package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
	"goproxy/domain/events"
	"log"
	"os"
	"time"
)

type TrafficCollector struct {
	redisClient *redis.Client
}

func NewTrafficCollector() (*TrafficCollector, error) {
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

	return &TrafficCollector{
		redisClient: redisClient,
	}, nil
}

func (t *TrafficCollector) ProcessEvents() {
	config, readConfigErr := t.readConfigMapFromEnv()
	if readConfigErr != nil {
		log.Fatal(readConfigErr)
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}
	defer func(consumer *kafka.Consumer) {
		_ = consumer.Close()
	}(consumer)

	err = consumer.SubscribeTopics([]string{os.Getenv("KAFKA_TOPIC")}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %s", err)
	}

	for {
		msg, readErr := consumer.ReadMessage(-1)
		if readErr == nil {
			t.consume(msg)
		} else {
			log.Printf("Consumer error: %v (%v)\n", readErr, msg)
		}
	}
}

func (t *TrafficCollector) readConfigMapFromEnv() (*kafka.ConfigMap, error) {
	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	groupId := os.Getenv("KAFKA_GROUP_ID")
	autoOffsetReset := os.Getenv("KAFKA_AUTO_OFFSET_RESET")
	topic := os.Getenv("KAFKA_TOPIC")

	if groupId == "" || autoOffsetReset == "" || topic == "" || bootstrapServers == "" {
		return nil, fmt.Errorf("invalid configuration")
	}

	return &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupId,
		"auto.offset.reset": autoOffsetReset,
	}, nil
}

func (t *TrafficCollector) consume(message *kafka.Message) {
	var event events.UserConsumedTrafficEvent
	err := json.Unmarshal(message.Value, &event)
	if err != nil {
		log.Printf("Invalid event: %v", err)
	}

	ctx := context.Background()
	key := fmt.Sprintf("user:%d:traffic:%s", event.UserId, time.Now().Format("02-01-2006"))

	currentTraffic, err := t.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		log.Printf("failed to get current traffic: %v", err)
		return
	}

	var inMb, outMb int
	if len(currentTraffic) > 0 {
		_, currentInScanErr := fmt.Sscanf(currentTraffic["inMb"], "%d", &inMb)
		if currentInScanErr != nil {
			inMb = 0
		}
		_, currentOutScanErr := fmt.Sscanf(currentTraffic["outMb"], "%d", &outMb)
		if currentOutScanErr != nil {
			outMb = 0
		}
	}

	inMb += event.InMb
	outMb += event.OutMb

	err = t.redisClient.HSet(ctx, key, map[string]interface{}{
		"inMb":  inMb,
		"outMb": outMb,
	}).Err()
	if err != nil {
		log.Printf("failed to update traffic: %v", err)
		return
	}

	err = t.redisClient.Expire(ctx, key, 24*time.Hour).Err()
	if err != nil {
		log.Printf("failed to set TTL: %v", err)
		return
	}
}
