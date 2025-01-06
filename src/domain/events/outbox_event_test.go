package events

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOutboxEvent_MarshalJSON(t *testing.T) {
	event, validationErr := NewOutboxEvent(1, "some_data", false, "test_event")
	if validationErr != nil {
		t.Fatalf("validation err: %s", validationErr)
	}

	jsonData, err := json.Marshal(event)

	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"Id": 1,
		"Payload": "some_data",
		"Published": false,
		"EventType": "test_event"
	}`, string(jsonData))
}

func TestOutboxEvent_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"Id": 1,
		"Payload": "some_data",
		"Published": false,
		"EventType": "test_event"
	}`

	var event OutboxEvent
	err := json.Unmarshal([]byte(jsonData), &event)

	assert.NoError(t, err)
	assert.Equal(t, 1, event.Id)
	assert.Equal(t, "some_data", event.Payload)
	assert.False(t, event.Published)
	assert.Equal(t, "test_event", event.EventType.Value())
}
