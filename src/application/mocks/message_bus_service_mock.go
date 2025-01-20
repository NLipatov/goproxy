package mocks

import (
	"errors"
	"goproxy/domain/events"
	"sync"
)

type MockMessageBusService struct {
	mu          sync.Mutex
	subscribed  []string
	eventsQueue []*events.OutboxEvent
	closed      bool
}

func NewMockMessageBusService() *MockMessageBusService {
	return &MockMessageBusService{
		eventsQueue: []*events.OutboxEvent{},
	}
}

func (m *MockMessageBusService) Subscribe(topics []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscribed = append(m.subscribed, topics...)
	return nil
}

func (m *MockMessageBusService) Consume() (*events.OutboxEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.eventsQueue) == 0 {
		return nil, errors.New("no events to consume")
	}

	event := m.eventsQueue[0]
	m.eventsQueue = m.eventsQueue[1:]
	return event, nil
}

func (m *MockMessageBusService) Produce(_ string, event events.OutboxEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.eventsQueue = append(m.eventsQueue, &event)
	return nil
}

func (m *MockMessageBusService) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

func (m *MockMessageBusService) WasCloseCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.closed
}
