package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	ticketsEntity "tickets/entities"
	ticketsEvent "tickets/message/event"
	ticketsOutbox "tickets/message/outbox"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
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

	return updateInTx(
		ctx,
		t.db,
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {
			availableSeats := 0
			err := tx.GetContext(
				ctx, &availableSeats, `
				SELECT
					number_of_tickets AS available_seats
				FROM
					shows
				WHERE
					show_id = $1
			`, booking.ShowID,
			)
			if err != nil {
				return fmt.Errorf("could not get available seats: %w", err)
			}

			alreadyBookedSeats := 0
			err = tx.GetContext(
				ctx, &alreadyBookedSeats, `
				SELECT
					COALESCE(SUM(number_of_tickets), 0) AS already_booked_seats
				FROM
					bookings
				WHERE
					show_id = $1
			`, booking.ShowID,
			)
			if err != nil {
				return fmt.Errorf("could not get already booked seats: %w", err)
			}

			if availableSeats-alreadyBookedSeats < booking.NumberOfTickets {
				// this is usually a bad idea, learn more here:
				// https://threedots.tech/post/introducing-clean-architecture/
				// we'll improve it later
				return echo.NewHTTPError(http.StatusBadRequest, "not enough seats available")
			}

			_, err = tx.NamedExecContext(
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

			outboxPublisher, err := ticketsOutbox.NewPublisherForDb(ctx, tx)
			if err != nil {
				return fmt.Errorf("could not create event bus: %w", err)
			}

			bus := ticketsEvent.NewEventBus(outboxPublisher, watermill.NewSlogLogger(log.FromContext(ctx)))
			return bus.Publish(
				ctx, ticketsEntity.BookingMade_v1{
					Header:          ticketsEntity.NewMessageHeader(),
					NumberOfTickets: booking.NumberOfTickets,
					BookingID:       booking.BookingID,
					CustomerEmail:   booking.CustomerEmail,
					ShowID:          booking.ShowID,
				},
			)
		},
	)
}
