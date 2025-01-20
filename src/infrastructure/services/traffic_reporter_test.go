package services

import (
	"context"
	"encoding/json"
	"goproxy/application/mocks"
	"goproxy/domain/events"
	"testing"
	"time"
)

func TestNewTrafficReporter(t *testing.T) {
	mockBus := mocks.NewMockMessageBusService()
	reporter := &TrafficReporter{
		buckets:    make(map[int]*TrafficBucket),
		messageBus: mockBus,
	}

	if reporter == nil {
		t.Fatal("TrafficReporter is nil")
	}
	if reporter.messageBus == nil {
		t.Fatal("MessageBusService is nil")
	}
}

func TestAddInBytesAndOutBytes(t *testing.T) {
	reporter := &TrafficReporter{
		buckets: make(map[int]*TrafficBucket),
	}

	reporter.AddInBytes(1, 100)
	reporter.AddOutBytes(1, 200)

	if reporter.buckets[1].InBytes != 100 {
		t.Errorf("Expected InBytes to be 100, got %d", reporter.buckets[1].InBytes)
	}
	if reporter.buckets[1].OutBytes != 200 {
		t.Errorf("Expected OutBytes to be 200, got %d", reporter.buckets[1].OutBytes)
	}
}

func TestFlushbuckets(t *testing.T) {
	testCtx, testCtxCancelFunc := context.WithCancel(context.Background())
	mockBus := mocks.NewMockMessageBusService()
	reporter := &TrafficReporter{
		buckets:    make(map[int]*TrafficBucket),
		eventQueue: make(chan events.UserConsumedTrafficEvent, 10),
		messageBus: mockBus,
	}

	go reporter.startEventWorker(testCtx)

	reporter.AddInBytes(1, 500)
	reporter.FlushBuckets()

	// wait for mockBus to process event queue
	time.Sleep(5 * time.Millisecond)

	event, err := mockBus.Consume()
	if err != nil {
		t.Fatalf("Failed to consume event: %v", err)
	}
	testCtxCancelFunc()

	if event == nil {
		t.Fatal("Expected event, got nil")
	}
	var consumedEvent events.UserConsumedTrafficEvent
	err = json.Unmarshal([]byte(event.Payload), &consumedEvent)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if consumedEvent.InBytes != 500 {
		t.Errorf("Expected InBytes to be 500, got %d", consumedEvent.InBytes)
	}
}

func TestProduceTrafficConsumedEvent(t *testing.T) {
	mockBus := mocks.NewMockMessageBusService()
	reporter := &TrafficReporter{
		messageBus: mockBus,
	}

	err := reporter.ProduceTrafficConsumedEvent(1, 100, 200)
	if err != nil {
		t.Fatalf("Failed to produce traffic event: %v", err)
	}

	event, err := mockBus.Consume()
	if err != nil {
		t.Fatalf("Failed to consume event: %v", err)
	}

	if event == nil {
		t.Fatal("Expected event, got nil")
	}

	var consumedEvent events.UserConsumedTrafficEvent
	err = json.Unmarshal([]byte(event.Payload), &consumedEvent)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if consumedEvent.UserId != 1 {
		t.Errorf("Expected UserId to be 1, got %d", consumedEvent.UserId)
	}
	if consumedEvent.InBytes != 100 {
		t.Errorf("Expected InBytes to be 100, got %d", consumedEvent.InBytes)
	}
	if consumedEvent.OutBytes != 200 {
		t.Errorf("Expected OutBytes to be 200, got %d", consumedEvent.OutBytes)
	}
}

func TestStop(t *testing.T) {
	reporter := &TrafficReporter{
		stopEventWorker: make(chan struct{}),
		eventQueue:      make(chan events.UserConsumedTrafficEvent),
	}

	go reporter.Stop()

	// wait for goroutine to call stop
	time.Sleep(1 * time.Millisecond)

	select {
	case <-reporter.stopEventWorker:
		// Success
	default:
		t.Error("Stop did not close stopEventWorker channel")
	}
}
