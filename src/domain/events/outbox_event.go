package events

import (
	"encoding/json"
	"goproxy/domain/valueobjects"
)

type OutboxEvent struct {
	Id        int
	Payload   string
	Published bool
	EventType valueobjects.OutboxEventType
}

func NewOutboxEvent(id int, payload string, published bool, eventType string) (OutboxEvent, error) {
	eType, eTypeErr := valueobjects.ParseEventTypeFromString(eventType)
	if eTypeErr != nil {
		return OutboxEvent{}, eTypeErr
	}

	return OutboxEvent{
		Id:        id,
		Payload:   payload,
		Published: published,
		EventType: eType,
	}, nil
}

func (o OutboxEvent) MarshalJSON() ([]byte, error) {
	type Alias OutboxEvent
	return json.Marshal(&struct {
		EventType string `json:"EventType"`
		Alias
	}{
		EventType: o.EventType.Value(),
		Alias:     (Alias)(o),
	})
}
