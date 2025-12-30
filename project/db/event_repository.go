package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	ticketsEntity "tickets/entities"
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

func (s EventsRepository) SaveEvent(ctx context.Context, event ticketsEntity.Event, eventName string, payload []byte) error

	_, err := s.db.NamedExecContext(
		ctx,
		`
       INSERT INTO
           shows (
                  show_id,
                  dead_nation_id,
                  number_of_tickets,
                  start_time,
                  title,
                  venue
           )
       VALUES
           (
             :show_id, 
             :dead_nation_id, 
             :number_of_tickets, 
             :start_time,
             :title,
             :venue
          )
       ON CONFLICT DO NOTHING`,
		show,
	)
	if err != nil {
		return fmt.Errorf("could not save show: %w", err)
	}

	return nil
}

func (s EventsRepository) ShowByID(ctx context.Context, showID string) (ticketsEntity.Show, error) {
	var show ticketsEntity.Show
	err := s.db.GetContext(ctx, &show, `SELECT * FROM shows WHERE show_id = $1`, showID)
	if err != nil {
		return ticketsEntity.Show{}, err
	}

	return show, nil
}
