package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	ticketsEntity "tickets/entities"
)

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) BookingRepository {
	if db == nil {
		panic("db is nil")
	}

	return BookingRepository{db: db}
}

func (t BookingRepository) AddBooking(ctx context.Context, booking ticketsEntity.Booking) error {
	_, err := t.db.NamedExecContext(
		ctx,
		`
       INSERT INTO
           bookings (
                  booking_id,
                  show_id,
                  number_of_tickets,
                  customer_email
           )
       VALUES
           (
             :booking_id, 
             :show_id, 
             :number_of_tickets, 
             :customer_email
          )
       ON CONFLICT DO NOTHING`,
		booking,
	)
	if err != nil {
		return fmt.Errorf("could not save booking: %w", err)
	}

	return nil
}
