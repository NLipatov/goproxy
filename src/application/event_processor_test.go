package application

import (
	"context"
	"errors"
	"goproxy/application/mocks"
	"goproxy/domain/events"
	"goproxy/domain/valueobjects"
	"testing"
	"time"
)

type MockEventHandler struct {
	handledEvents []string
	returnError   bool
}

func (m *MockEventHandler) Handle(payload string) error {
	m.handledEvents = append(m.handledEvents, payload)
	if m.returnError {
		return errors.New("mock handler error")
	}
	return nil
}

func TestEventProcessor_Start(t *testing.T) {
	// Arrange
	mockBus := mocks.NewMockMessageBusService()
	mockHandler := &MockEventHandler{}
	processor := NewEventProcessor(mockBus).
		RegisterHandler("TestEvent", mockHandler).
		RegisterTopic("TestTopic")

	if err := processor.Build(); err != nil {
		t.Fatalf("Failed to build EventProcessor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventType, eventTypeErr := valueobjects.ParseEventTypeFromString("TestEvent")
	if eventTypeErr != nil {
		t.Fatalf("Failed to parse event type from payload: %v", eventTypeErr)
	}

	_ = mockBus.Produce("TestTopic", events.OutboxEvent{
		EventType: eventType,
		Payload:   "test_payload",
	})

	// Act
	go func() {
		if err := processor.Start(ctx); err != nil {
			t.Errorf("Failed to start EventProcessor: %v", err)
		}
	}()

	// Wait for the event to be processed
	time.Sleep(100 * time.Millisecond)

	// Assert
	if len(mockHandler.handledEvents) != 1 {
		t.Fatalf("Expected 1 handled event, got %d", len(mockHandler.handledEvents))
	}
	if mockHandler.handledEvents[0] != "test_payload" {
		t.Errorf("Unexpected payload handled: %s", mockHandler.handledEvents[0])
	}
}

func TestEventProcessor_MissingHandler(t *testing.T) {
	// Arrange
	mockBus := mocks.NewMockMessageBusService()
	processor := NewEventProcessor(mockBus).
		RegisterHandler("SomeEvent", &MockEventHandler{}).
		RegisterTopic("TestTopic")

	if err := processor.Build(); err != nil {
		t.Fatalf("Failed to build EventProcessor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventType, eventTypeErr := valueobjects.ParseEventTypeFromString("UnregisteredEvent")
	if eventTypeErr != nil {
		t.Fatalf("Failed to parse event type from payload: %v", eventTypeErr)
	}

	_ = mockBus.Produce("TestTopic", events.OutboxEvent{
		EventType: eventType,
		Payload:   "test_payload",
	})

	// Act
	go func() {
		if err := processor.Start(ctx); err != nil {
			t.Errorf("Failed to start EventProcessor: %v", err)
		}
	}()

	// Wait for the event to be processed
	time.Sleep(100 * time.Millisecond)

	// Assert
	// No handlers registered for "UnregisteredEvent", so nothing should happen
}

func TestEventProcessor_Shutdown(t *testing.T) {
	// Arrange
	mockBus := mocks.NewMockMessageBusService()
	mockHandler := &MockEventHandler{}
	processor := NewEventProcessor(mockBus).
		RegisterHandler("TestEvent", mockHandler).
		RegisterTopic("TestTopic")

	if err := processor.Build(); err != nil {
		t.Fatalf("Failed to build EventProcessor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Act
	go func() {
		if err := processor.Start(ctx); err != nil {
			t.Errorf("Failed to start EventProcessor: %v", err)
		}
	}()

	// Simulate shutdown
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Assert
	time.Sleep(100 * time.Millisecond) // Give time for graceful shutdown
	if !mockBus.WasCloseCalled() {
		t.Errorf("MessageBus was not closed during shutdown")
	}
}
