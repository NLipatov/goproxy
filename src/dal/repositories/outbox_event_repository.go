package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/domain/events"
)

type DomainEventRepository struct {
	db *sql.DB
}

func NewDomainEventRepository(db *sql.DB) (*DomainEventRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &DomainEventRepository{db: db}, nil
}

const getByIdQuery = `
	SELECT id, payload, published FROM public.outbox 
    WHERE id = $1
`

const createQuery = `
	INSERT INTO public.outbox (payload) VALUES ($1) RETURNING id
`

const updateQuery = `
	UPDATE public.outbox SET payload = $1 WHERE id = $2
`

const deleteQuery = `DELETE FROM public.outbox WHERE id = $1`

func (d DomainEventRepository) GetById(id int) (events.OutboxEvent, error) {
	var payload string
	var published bool
	var resultId int

	err := d.db.
		QueryRow(getByIdQuery, id).
		Scan(&resultId, &payload, &published)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return events.OutboxEvent{}, fmt.Errorf("event not found: %v", err)
		}
		return events.OutboxEvent{}, fmt.Errorf("could not load event: %v", err)
	}

	return events.NewOutboxEvent(resultId, payload, published), nil
}

func (d DomainEventRepository) Create(event events.OutboxEvent) (int, error) {
	var id int
	err := d.db.QueryRow(createQuery, event.Payload).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("could not create event: %v", err)
	}
	return id, nil
}

func (d DomainEventRepository) Update(event events.OutboxEvent) error {
	result, err := d.db.Exec(updateQuery, event.Payload, event.Id)
	if err != nil {
		return fmt.Errorf("could not update event: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("could not update event: no rows updated")
	}

	return nil
}

func (d DomainEventRepository) Delete(event events.OutboxEvent) error {
	result, err := d.db.Exec(deleteQuery, event.Id)
	if err != nil {
		return fmt.Errorf("could not delete event: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
