package application

import (
	"context"
	"errors"
	"fmt"
	"log"
)

type EventHandler interface {
	Handle(payload string) error
}

type EventProcessor struct {
	messageBus MessageBusService
	handlerMap map[string]EventHandler
	topics     []string
}

func NewEventProcessor(messageBus MessageBusService) *EventProcessor {
	return &EventProcessor{
		messageBus: messageBus,
		handlerMap: make(map[string]EventHandler),
	}
}

func (e *EventProcessor) RegisterHandler(eventType string, handler EventHandler) *EventProcessor {
	e.handlerMap[eventType] = handler
	return e
}

func (e *EventProcessor) RegisterTopic(topic string) *EventProcessor {
	e.topics = append(e.topics, topic)
	return e
}

func (e *EventProcessor) Build() error {
	if e.topics == nil || len(e.topics) == 0 {
		return errors.New("topics is empty")
	}

	if len(e.handlerMap) == 0 {
		return errors.New("no handlers registered")
	}

	return nil
}

func (e *EventProcessor) Start(ctx context.Context) error {
	if err := e.messageBus.Subscribe(e.topics); err != nil {
		log.Fatalf("Failed to subscribe to topics: %v", err)
	}

	go func() {
		defer func(messageBus MessageBusService) {
			_ = messageBus.Close()
		}(e.messageBus)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				eventProcessingErr := e.ProcessNextEvent()
				if eventProcessingErr != nil {
					log.Printf("failed to process event: %v", eventProcessingErr)
				}
			}
		}
	}()

	return nil
}

func (e *EventProcessor) ProcessNextEvent() error {
	event, err := e.messageBus.Consume()
	if err != nil {
		return fmt.Errorf("failed to consume event: %v", err)
	}

	eventHandler, ok := e.handlerMap[event.EventType.Value()]
	if !ok {
		// If there's no handlers in this processor - skip event.
		// Other processor probably will handle this event.
		return nil
	}

	return eventHandler.Handle(event.Payload)
}
