package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	ticketsEntity "tickets/entities"
)

type OpsBookingReadModel struct {
	db       *sqlx.DB
	eventBus *cqrs.EventBus
}

func NewOpsBookingReadModel(db *sqlx.DB, eventBus *cqrs.EventBus) OpsBookingReadModel {
	if db == nil {
		panic("db is nil")
	}

	return OpsBookingReadModel{db: db, eventBus: eventBus}
}

func (r OpsBookingReadModel) AllBookingsByDate(date string) ([]ticketsEntity.OpsBooking, error) {
	if date == "" {
		return r.AllBookings()
	}
	query := `SELECT payload FROM read_model_ops_bookings WHERE booking_id IN (
				SELECT booking_id FROM (
					SELECT booking_id,
						DATE(jsonb_path_query(payload, '$.tickets.*.receipt_issued_at')::text) as receipt_issued_at
					FROM
						read_model_ops_bookings
				) bookings_within_date
				WHERE receipt_issued_at = $1
			)`

	var queryArgs []any
	queryArgs = append(queryArgs, date)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ticketsEntity.OpsBooking
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}

		var reservation ticketsEntity.OpsBooking
		if err := json.Unmarshal(payload, &reservation); err != nil {
			return nil, err
		}

		result = append(result, reservation)
	}

	return result, nil
}

func (r OpsBookingReadModel) AllBookings() ([]ticketsEntity.OpsBooking, error) {
	query := "SELECT payload FROM read_model_ops_bookings"
	var queryArgs []any

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ticketsEntity.OpsBooking
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}

		var reservation ticketsEntity.OpsBooking
		if err := json.Unmarshal(payload, &reservation); err != nil {
			return nil, err
		}

		result = append(result, reservation)
	}

	return result, nil
}

func (r OpsBookingReadModel) ReservationReadModel(
	ctx context.Context,
	bookingID string,
) (ticketsEntity.OpsBooking, error) {
	return r.findReadModelByBookingID(ctx, bookingID, r.db)
}

func (r OpsBookingReadModel) OnBookingMade(ctx context.Context, bookingMade *ticketsEntity.BookingMade_v1) error {
	// this is the first event that should arrive, so we create the read model
	bookingID, err := uuid.Parse(bookingMade.BookingID)
	if err != nil {
		return err
	}
	err = r.createReadModel(
		ctx, ticketsEntity.OpsBooking{
			BookingID:  bookingID,
			Tickets:    nil,
			LastUpdate: time.Now(),
			BookedAt:   bookingMade.Header.PublishedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("could not create read model: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) OnTicketBookingConfirmed(
	ctx context.Context,
	event *ticketsEntity.TicketBookingConfirmed_v1,
) error {
	return r.updateReadModelByBookingID(
		ctx,
		event.BookingID,
		func(rm ticketsEntity.OpsBooking) (ticketsEntity.OpsBooking, error) {

			ticket, ok := rm.Tickets[event.TicketID]
			if !ok {
				// we are using zero-value of OpsTicket
				log.
					FromContext(ctx).
					With("ticket_id", event.TicketID).
					Debug("Creating ticket read model for ticket %s")
			}

			ticket.PriceAmount = event.Price.Amount
			ticket.PriceCurrency = event.Price.Currency
			ticket.CustomerEmail = event.CustomerEmail
			ticket.ConfirmedAt = event.Header.PublishedAt

			rm.Tickets[event.TicketID] = ticket

			return rm, nil
		},
	)
}

func (r OpsBookingReadModel) OnTicketRefunded(ctx context.Context, event *ticketsEntity.TicketRefunded_v1) error {
	return r.updateReadModelByTicketID(
		ctx,
		event.TicketID,
		func(rm ticketsEntity.OpsTicket) (ticketsEntity.OpsTicket, error) {
			rm.RefundedAt = event.Header.PublishedAt

			return rm, nil
		},
	)
}

func (r OpsBookingReadModel) OnTicketPrinted(ctx context.Context, event *ticketsEntity.TicketPrinted_v1) error {
	return r.updateReadModelByTicketID(
		ctx,
		event.TicketID,
		func(rm ticketsEntity.OpsTicket) (ticketsEntity.OpsTicket, error) {
			rm.PrintedAt = event.Header.PublishedAt
			rm.PrintedFileName = event.FileName

			return rm, nil
		},
	)
}

func (r OpsBookingReadModel) OnTicketReceiptIssued(
	ctx context.Context,
	issued *ticketsEntity.TicketReceiptIssued_v1,
) error {
	return r.updateReadModelByTicketID(
		ctx,
		issued.TicketID,
		func(rm ticketsEntity.OpsTicket) (ticketsEntity.OpsTicket, error) {
			rm.ReceiptIssuedAt = issued.IssuedAt
			rm.ReceiptNumber = issued.ReceiptNumber

			return rm, nil
		},
	)
}

func (r OpsBookingReadModel) createReadModel(
	ctx context.Context,
	booking ticketsEntity.OpsBooking,
) (err error) {
	payload, err := json.Marshal(booking)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx, `
		INSERT INTO 
		    read_model_ops_bookings (payload, booking_id)
		VALUES
			($1, $2)
		ON CONFLICT (booking_id) DO NOTHING; -- read model may be already updated by another event - we don't want to override
`, payload, booking.BookingID,
	)

	if err != nil {
		return fmt.Errorf("could not create read model: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) updateReadModelByBookingID(
	ctx context.Context,
	bookingID string,
	updateFunc func(ticket ticketsEntity.OpsBooking) (ticketsEntity.OpsBooking, error),
) (err error) {
	err = updateInTx(
		ctx,
		r.db,
		sql.LevelRepeatableRead,
		func(ctx context.Context, tx *sqlx.Tx) error {
			rm, err := r.findReadModelByBookingID(ctx, bookingID, tx)
			if errors.Is(err, sql.ErrNoRows) {
				// events arrived out of order - it should spin until the read model is created
				return fmt.Errorf("read model for booking %s not exist yet", bookingID)
			} else if err != nil {
				return fmt.Errorf("could not find read model: %w", err)
			}

			updatedRm, err := updateFunc(rm)
			if err != nil {
				return err
			}

			return r.updateReadModel(ctx, tx, updatedRm)
		},
	)
	if err != nil {
		return err
	}

	return r.eventBus.Publish(
		ctx, &ticketsEntity.InternalOpsReadModelUpdated{
			Header:    ticketsEntity.NewMessageHeader(),
			BookingID: uuid.MustParse(bookingID),
		},
	)
}

func (r OpsBookingReadModel) updateReadModelByTicketID(
	ctx context.Context,
	ticketID string,
	updateFunc func(ticket ticketsEntity.OpsTicket) (ticketsEntity.OpsTicket, error),
) (err error) {
	return updateInTx(
		ctx,
		r.db,
		sql.LevelRepeatableRead,
		func(ctx context.Context, tx *sqlx.Tx) error {
			rm, err := r.findReadModelByTicketID(ctx, ticketID, tx)
			if errors.Is(err, sql.ErrNoRows) {
				// events arrived out of order - it should spin until the read model is created
				return fmt.Errorf("read model for ticket %s not exist yet", ticketID)
			} else if err != nil {
				return fmt.Errorf("could not find read model: %w", err)
			}

			ticket, _ := rm.Tickets[ticketID]

			updatedRm, err := updateFunc(ticket)
			if err != nil {
				return err
			}

			rm.Tickets[ticketID] = updatedRm

			return r.updateReadModel(ctx, tx, rm)
		},
	)
}

func (r OpsBookingReadModel) updateReadModel(
	ctx context.Context,
	tx *sqlx.Tx,
	rm ticketsEntity.OpsBooking,
) error {
	rm.LastUpdate = time.Now()

	payload, err := json.Marshal(rm)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx, `
		INSERT INTO 
			read_model_ops_bookings (payload, booking_id)
		VALUES
			($1, $2)
		ON CONFLICT (booking_id) DO UPDATE SET payload = excluded.payload;
		`, payload, rm.BookingID,
	)
	if err != nil {
		return fmt.Errorf("could not update read model: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) findReadModelByTicketID(
	ctx context.Context,
	ticketID string,
	db dbExecutor,
) (ticketsEntity.OpsBooking, error) {
	var payload []byte

	err := db.QueryRowContext(
		ctx,
		"SELECT payload FROM read_model_ops_bookings WHERE payload::jsonb -> 'tickets' ? $1",
		ticketID,
	).Scan(&payload)
	if err != nil {
		return ticketsEntity.OpsBooking{}, err
	}

	return r.unmarshalReadModelFromDB(payload)
}

func (r OpsBookingReadModel) findReadModelByBookingID(
	ctx context.Context,
	bookingID string,
	db dbExecutor,
) (ticketsEntity.OpsBooking, error) {
	var payload []byte

	err := db.QueryRowContext(
		ctx,
		"SELECT payload FROM read_model_ops_bookings WHERE booking_id = $1",
		bookingID,
	).Scan(&payload)
	if err != nil {
		return ticketsEntity.OpsBooking{}, err
	}

	return r.unmarshalReadModelFromDB(payload)
}

func (r OpsBookingReadModel) unmarshalReadModelFromDB(payload []byte) (ticketsEntity.OpsBooking, error) {
	var dbReadModel ticketsEntity.OpsBooking
	if err := json.Unmarshal(payload, &dbReadModel); err != nil {
		return ticketsEntity.OpsBooking{}, err
	}

	if dbReadModel.Tickets == nil {
		dbReadModel.Tickets = map[string]ticketsEntity.OpsTicket{}
	}

	return dbReadModel, nil
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
