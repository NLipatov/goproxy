package services

import (
	"github.com/stretchr/testify/mock"
	"goproxy/domain/events"
)

type MockMessageBusService struct {
	mock.Mock
}

func (m *MockMessageBusService) Subscribe(topics []string) error {
	args := m.Called(topics)
	return args.Error(0)
}

func (m *MockMessageBusService) Consume() (*events.OutboxEvent, error) {
	args := m.Called()
	return args.Get(0).(*events.OutboxEvent), args.Error(1)
}

func (m *MockMessageBusService) Produce(topic string, event events.OutboxEvent) error {
	args := m.Called(topic, event)
	return args.Error(0)
}

func (m *MockMessageBusService) Close() error {
	args := m.Called()
	return args.Error(0)
}
