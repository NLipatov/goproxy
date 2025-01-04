package infrastructure

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/domain/events"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"time"
)

type userTraffic struct {
	InBytes  int
	OutBytes int
}

type TrafficCollector struct {
	cache      application.CacheWithTTL[userTraffic]
	messageBus application.MessageBusService
}

func NewTrafficCollector() (*TrafficCollector, error) {
	messageBus, err := instantiateMessageBusService()
	if err != nil {
		return nil, err
	}

	redisCache, redisCacheClientErr := services.NewRedisCache[userTraffic]()
	if redisCacheClientErr != nil {
		log.Fatal(redisCacheClientErr)
	}

	return &TrafficCollector{
		cache:      redisCache,
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

	key := fmt.Sprintf("user:%d:traffic:%s", event.UserId, time.Now().Format("02-01-2006"))

	currentTraffic, err := t.cache.Get(key)
	if err != nil {
		log.Printf("failed to get current traffic: %v", err)
		return
	}

	currentTraffic.InBytes += event.InBytes
	currentTraffic.OutBytes += event.OutBytes

	err = t.cache.Set(key, currentTraffic)
	if err != nil {
		log.Printf("failed to update traffic: %v", err)
		return
	}

	err = t.cache.Expire(key, 24*time.Hour)
	if err != nil {
		log.Printf("failed to set TTL: %v", err)
		return
	}
}
