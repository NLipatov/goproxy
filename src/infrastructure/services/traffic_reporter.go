package services

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/domain/events"
	"log"
	"os"
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

func NewTrafficReporter(userId int, threshold int64, interval time.Duration) *TrafficReporter {
	messageBusService, err := instantiateMessageBusService()
	if err != nil {
		log.Fatal(err)
	}

	return &TrafficReporter{
		userId:         userId,
		thresholdBytes: threshold,
		interval:       interval,
		lastSent:       time.Now(),
		messageBus:     messageBusService,
	}
}

func instantiateMessageBusService() (application.MessageBusService, error) {
	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	groupId := os.Getenv("TC_KAFKA_GROUP_ID")
	autoOffsetReset := os.Getenv("TC_KAFKA_AUTO_OFFSET_RESET")
	topic := os.Getenv("TC_KAFKA_TOPIC")

	if groupId == "" || autoOffsetReset == "" || topic == "" || bootstrapServers == "" {
		return nil, fmt.Errorf("invalid configuration")
	}

	return NewKafkaService(bootstrapServers, groupId, autoOffsetReset)
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
		tr.lastSent = time.Now()
		return
	}
	_ = tr.ProduceTrafficConsumedEvent(int(in), int(out))

	atomic.StoreInt64(&tr.inBytes, 0)
	atomic.StoreInt64(&tr.outBytes, 0)
	tr.lastSent = time.Now()
}

func (tr *TrafficReporter) SendFinal() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	in := atomic.LoadInt64(&tr.inBytes)
	out := atomic.LoadInt64(&tr.outBytes)
	if in == 0 && out == 0 {
		return
	}
	_ = tr.ProduceTrafficConsumedEvent(int(in), int(out))
}

func (tr *TrafficReporter) ProduceTrafficConsumedEvent(in, out int) error {
	event := events.NewUserConsumedTrafficEvent(tr.userId, in, out)
	eventJson, serializationErr := json.Marshal(event)
	if serializationErr != nil {
		log.Printf("Could not serialize consumed traffic event: %v", serializationErr)
		return serializationErr
	}

	outboxEvent := events.NewOutboxEvent(0, string(eventJson), false)
	produceErr := tr.messageBus.Produce("user-traffic", outboxEvent)
	if produceErr != nil {
		log.Printf("Could not produce outbox event: %v", produceErr)
		return produceErr
	}
	return nil
}
