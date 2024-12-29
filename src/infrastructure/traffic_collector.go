package infrastructure

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"goproxy/domain/events"
	"log"
	"os"
	"sync"
)

type TrafficCollector struct {
	mu         sync.Mutex
	userEvents map[int][]events.UserConsumedTrafficEvent
}

func NewTrafficCollector() (*TrafficCollector, error) {
	return &TrafficCollector{
		mu:         sync.Mutex{},
		userEvents: make(map[int][]events.UserConsumedTrafficEvent),
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
	var userConsumedTrafficEvent events.UserConsumedTrafficEvent
	err := json.Unmarshal(message.Value, &userConsumedTrafficEvent)
	if err != nil {
		log.Printf("Invalid event: %v", err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	val, contains := t.userEvents[userConsumedTrafficEvent.UserId]
	if contains {
		t.userEvents[userConsumedTrafficEvent.UserId] = append(val, userConsumedTrafficEvent)
	} else {
		t.userEvents[userConsumedTrafficEvent.UserId] = []events.UserConsumedTrafficEvent{
			userConsumedTrafficEvent,
		}
	}
}
