package services

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/domain"
	"goproxy/domain/events"
	"goproxy/infrastructure/config"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type TrafficReporter struct {
	userId         int
	inBytes        int64
	outBytes       int64
	thresholdBytes int64
	interval       time.Duration

	mu         sync.Mutex
	lastSent   time.Time
	messageBus application.MessageBusService
}

func NewTrafficReporter(userId int, threshold int64, interval time.Duration) (*TrafficReporter, error) {
	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(domain.PROXY)
	if kafkaConfigErr != nil {
		log.Fatalf("Error creating kafka config: %v", kafkaConfigErr)
	}

	messageBusService, err := NewKafkaService(kafkaConfig)
	if err != nil {
		return nil, err
	}

	return &TrafficReporter{
		userId:         userId,
		thresholdBytes: threshold,
		interval:       interval,
		lastSent:       time.Now().UTC(),
		messageBus:     messageBusService,
	}, nil
}

func (tr *TrafficReporter) AddInBytes(n int64) {
	if n == 0 {
		return
	}
	atomic.AddInt64(&tr.inBytes, n)
	tr.checkAndSend()
}

func (tr *TrafficReporter) AddOutBytes(n int64) {
	if n == 0 {
		return
	}
	atomic.AddInt64(&tr.outBytes, n)
	tr.checkAndSend()
}

func (tr *TrafficReporter) checkAndSend() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	in := atomic.LoadInt64(&tr.inBytes)
	out := atomic.LoadInt64(&tr.outBytes)

	if in+out >= tr.thresholdBytes || time.Since(tr.lastSent) >= tr.interval {
		tr.SendIntermediate(in, out)
	}
}

func (tr *TrafficReporter) SendIntermediate(in, out int64) {
	if in == 0 && out == 0 {
		tr.lastSent = time.Now().UTC()
		return
	}
	_ = tr.ProduceTrafficConsumedEvent(in, out)

	atomic.StoreInt64(&tr.inBytes, 0)
	atomic.StoreInt64(&tr.outBytes, 0)
	tr.lastSent = time.Now().UTC()
}

func (tr *TrafficReporter) SendFinal() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	in := atomic.LoadInt64(&tr.inBytes)
	out := atomic.LoadInt64(&tr.outBytes)
	if in == 0 && out == 0 {
		return
	}
	_ = tr.ProduceTrafficConsumedEvent(in, out)
}

func (tr *TrafficReporter) ProduceTrafficConsumedEvent(in, out int64) error {
	event := events.NewUserConsumedTrafficEvent(tr.userId, in, out)
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
