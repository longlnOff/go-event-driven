package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	ticketsEntity "tickets/entities"
	"github.com/lib/pq"
)

type EventsRepository struct {
	db *sqlx.DB
}

func NewEventsRepository(db *sqlx.DB) EventsRepository {
	if db == nil {
		panic("db is nil")
	}

	return EventsRepository{db: db}
}

func (s EventsRepository) SaveEvent(
	ctx context.Context,
	event ticketsEntity.Event,
	eventName string,
	payload []byte,
) error {
	_, err := s.db.ExecContext(
		ctx,
		`
        INSERT INTO events (
            event_id,
            published_at,
            event_name,
            event_payload
        )
        VALUES ($1, $2, $3, $4)
        ON CONFLICT DO NOTHING`,
		event.Header.ID,
		event.Header.PublishedAt,
		eventName,
		payload,
	)
	var postgresError *pq.Error
	if errors.As(err, &postgresError) && postgresError.Code.Name() == "unique_violation" {
		// handling re-delivery
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not store %s event in data lake: %w", event.Header.ID, err)
	}
	return nil

}
