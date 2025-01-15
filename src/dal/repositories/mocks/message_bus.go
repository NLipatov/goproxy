package mocks

import (
	"errors"
	"goproxy/domain/events"
	"sync"
)

type MockMessageBusService struct {
	topics    map[string][]events.OutboxEvent
	mutex     sync.RWMutex
	isClosed  bool
	consumeCh chan events.OutboxEvent
}

func NewMockMessageBusService() *MockMessageBusService {
	return &MockMessageBusService{
		topics:    make(map[string][]events.OutboxEvent),
		consumeCh: make(chan events.OutboxEvent, 100),
	}
}

func (m *MockMessageBusService) Subscribe(topics []string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isClosed {
		return errors.New("message bus is closed")
	}

	for _, topic := range topics {
		if _, exists := m.topics[topic]; !exists {
			m.topics[topic] = []events.OutboxEvent{}
		}
	}

	return nil
}

func (m *MockMessageBusService) Consume() (*events.OutboxEvent, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.isClosed {
		return nil, errors.New("message bus is closed")
	}

	event, ok := <-m.consumeCh
	if !ok {
		return nil, errors.New("no events to consume")
	}

	return &event, nil
}

func (m *MockMessageBusService) Produce(topic string, event events.OutboxEvent) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isClosed {
		return errors.New("message bus is closed")
	}

	if _, exists := m.topics[topic]; !exists {
		return errors.New("topic not found")
	}

	m.topics[topic] = append(m.topics[topic], event)

	select {
	case m.consumeCh <- event:
	default:
	}

	return nil
}

func (m *MockMessageBusService) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isClosed {
		return errors.New("message bus is already closed")
	}

	m.isClosed = true
	close(m.consumeCh)
	return nil
}
