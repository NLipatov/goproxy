package services

import (
	"context"
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/domain"
	"goproxy/domain/events"
	"goproxy/infrastructure/config"
	"log"
	"sync"
)

type TrafficReporter struct {
	mu              sync.RWMutex
	buckets         map[int]*TrafficBucket
	messageBus      application.MessageBusService
	eventQueue      chan events.UserConsumedTrafficEvent
	stopEventWorker chan struct{}
}

type TrafficBucket struct {
	InBytes  int64
	OutBytes int64
}

func NewTrafficReporter() (*TrafficReporter, error) {
	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(domain.PROXY)
	if kafkaConfigErr != nil {
		log.Fatalf("Error creating kafka config: %v", kafkaConfigErr)
	}

	messageBusService, err := NewKafkaService(kafkaConfig)
	if err != nil {
		return nil, err
	}

	reporter := &TrafficReporter{
		buckets:         make(map[int]*TrafficBucket),
		messageBus:      messageBusService,
		eventQueue:      make(chan events.UserConsumedTrafficEvent, 100), // Буфер для очереди событий
		stopEventWorker: make(chan struct{}),
	}

	go reporter.startEventWorker(context.Background())
	return reporter, nil
}

func (tr *TrafficReporter) AddInBytes(userId int, n int64) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if n == 0 {
		return
	}
	bucket, exists := tr.buckets[userId]
	if !exists {
		bucket = &TrafficBucket{}
		tr.buckets[userId] = bucket
	}

	bucket.InBytes += n
}

func (tr *TrafficReporter) AddOutBytes(userId int, n int64) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	bucket, exists := tr.buckets[userId]
	if !exists {
		bucket = &TrafficBucket{}
		tr.buckets[userId] = bucket
	}
	bucket.OutBytes += n
}
func (tr *TrafficReporter) startEventWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-tr.eventQueue:
			eventJson, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to serialize event: %v", err)
				continue
			}

			outboxEvent, err := events.NewOutboxEvent(0, string(eventJson), false, "UserConsumedTrafficEvent")
			if err != nil {
				log.Printf("Failed to create outbox event: %v", err)
				continue
			}

			if err := tr.messageBus.Produce(fmt.Sprintf("%s", domain.PLAN), outboxEvent); err != nil {
				log.Printf("Failed to produce event: %v", err)
			}
		case <-tr.stopEventWorker:
			return
		}
	}
}

func (tr *TrafficReporter) FlushBuckets() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	for userId, bucket := range tr.buckets {
		if bucket.InBytes > 0 || bucket.OutBytes > 0 {
			tr.eventQueue <- events.NewUserConsumedTrafficEvent(userId, bucket.InBytes, bucket.OutBytes)
			bucket.InBytes = 0
			bucket.OutBytes = 0
		}
	}
}

func (tr *TrafficReporter) Stop() {
	close(tr.stopEventWorker)
	close(tr.eventQueue)
}

func (tr *TrafficReporter) ProduceTrafficConsumedEvent(userId int, in, out int64) error {
	event := events.NewUserConsumedTrafficEvent(userId, in, out)
	eventJson, serializationErr := json.Marshal(event)
	if serializationErr != nil {
		log.Printf("Could not serialize consumed traffic event: %v", serializationErr)
		return serializationErr
	}

	outboxEvent, outboxEventValidationErr := events.NewOutboxEvent(0, string(eventJson), false, "UserConsumedTrafficEvent")
	if outboxEventValidationErr != nil {
		return outboxEventValidationErr
	}

	produceErr := tr.messageBus.Produce(fmt.Sprintf("%s", domain.PLAN), outboxEvent)
	if produceErr != nil {
		log.Printf("Could not produce outbox event: %v", produceErr)
		return produceErr
	}
	return nil
}
