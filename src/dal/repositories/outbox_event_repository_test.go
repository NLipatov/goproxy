package repositories

import (
	"database/sql"
	_ "github.com/lib/pq"
	"goproxy/domain/events"
	"testing"
)

func TestDomainEventRepository(t *testing.T) {
	db, cleanup := prepareDb(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	repo, err := NewDomainEventRepository(db)
	assertNoError(t, err, "Failed to create DomainEventRepository")

	t.Run("GetById", func(t *testing.T) {
		eventId := insertTestEvent(repo, t)
		event, err := repo.GetById(eventId)
		assertNoError(t, err, "Failed to load event by Id")
		assertEventExists(t, event, eventId)
	})

	t.Run("Insert", func(t *testing.T) {
		eventId := insertTestEvent(repo, t)
		event, err := repo.GetById(eventId)
		assertNoError(t, err, "Failed to load inserted event")
		assertEventExists(t, event, eventId)
	})

	t.Run("Update", func(t *testing.T) {
		eventId := insertTestEvent(repo, t)
		event, err := repo.GetById(eventId)
		assertNoError(t, err, "Failed to load event by Id")

		updatedPayload := `{"type":"UPDATED_TEST_EVENT","data":{"key":"updated-value"}}`
		event.Payload = updatedPayload

		assertNoError(t, repo.Update(event), "Failed to update event")
		updatedEvent, err := repo.GetById(eventId)
		assertNoError(t, err, "Failed to load updated event")

		assertJSONEqual(t, updatedPayload, updatedEvent.Payload)
	})

	t.Run("Delete", func(t *testing.T) {
		eventId := insertTestEvent(repo, t)
		event, err := repo.GetById(eventId)
		assertNoError(t, err, "Failed to load event by Id")

		assertNoError(t, repo.Delete(event), "Failed to delete event")
		_, err = repo.GetById(eventId)
		if err == nil {
			t.Fatalf("Expected error when loading deleted event, but got nil")
		}
	})
}

func insertTestEvent(repo *DomainEventRepository, t *testing.T) int {
	payload := `{"type":"TEST_EVENT","data":{"key":"value"}}`
	event := events.OutboxEvent{
		Payload:   payload,
		Published: false,
	}
	id, err := repo.Create(event)
	assertNoError(t, err, "Failed to insert test event")
	return id
}

func assertEventExists(t *testing.T, event events.OutboxEvent, expectedId int) {
	if event.Id != expectedId {
		t.Errorf("Expected event ID %d, got %d", expectedId, event.Id)
	}
	if event.Payload == "" {
		t.Errorf("Expected non-empty payload, but got empty")
	}
}
