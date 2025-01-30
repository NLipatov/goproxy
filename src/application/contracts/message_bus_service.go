package contracts

import (
	"goproxy/domain/events"
)

type MessageBusService interface {
	Subscribe(topics []string) error
	Consume() (*events.OutboxEvent, error)
	Produce(topic string, event events.OutboxEvent) error
	Close() error
}
