package valueobjects

import (
	"encoding/json"
	"fmt"
)

type OutboxEventType struct {
	value string
}

func ParseEventTypeFromString(eventType string) (OutboxEventType, error) {
	if len(eventType) > 100 {
		return OutboxEventType{}, fmt.Errorf("event type max length is 100 characters")
	}

	return OutboxEventType{
		value: eventType,
	}, nil
}

func (o *OutboxEventType) Value() string {
	return o.value
}

func (o *OutboxEventType) UnmarshalJSON(data []byte) error {
	var eventType string
	if err := json.Unmarshal(data, &eventType); err != nil {
		return err
	}

	parsed, err := ParseEventTypeFromString(eventType)
	if err != nil {
		return err
	}

	*o = parsed
	return nil
}
