package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	ticketsEntity "tickets/entities"
)

type ShowsRepository struct {
	db *sqlx.DB
}

func NewShowsRepository(db *sqlx.DB) ShowsRepository {
	if db == nil {
		panic("db is nil")
	}

	return ShowsRepository{db: db}
}

func (t ShowsRepository) AddShow(ctx context.Context, show ticketsEntity.Show) error {
	_, err := t.db.NamedExecContext(
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
