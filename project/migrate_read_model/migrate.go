package migrate_read_model

import (
	"context"
	"encoding/json"
	"fmt"
	ticketsDB "tickets/db"
	ticketsEntity "tickets/entities"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

type EventsRepository interface {
	GetEvents(ctx context.Context) ([]ticketsEntity.DataLakeEvent, error)
}

func unmarshalDataLakeEvent[T any](event ticketsEntity.DataLakeEvent) (*T, error) {
	eventInstance := new(T)

	err := json.Unmarshal(event.EventPayload, &eventInstance)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event %s: %w", event.EventName, err)
	}

	return eventInstance, nil
}

func migrateEvent(ctx context.Context, event ticketsEntity.DataLakeEvent, rm ticketsDB.OpsBookingReadModel) error {
	switch event.EventName {
	case "BookingMade_v0":
		bookingMade, err := unmarshalDataLakeEvent[bookingMade_v0](event)
		if err != nil {
			return err
		}

		return rm.OnBookingMade(
			ctx, &ticketsEntity.BookingMade_v1{
				// you should map v0 event to your v1 event here
				Header:          bookingMade.Header,
				NumberOfTickets: bookingMade.NumberOfTickets,
				BookingID:       bookingMade.BookingID.String(),
				CustomerEmail:   bookingMade.CustomerEmail,
				ShowID:          bookingMade.ShowID.String(),
			},
		)
	case "TicketBookingConfirmed_v0":
		bookingConfirmedEvent, err := unmarshalDataLakeEvent[ticketBookingConfirmed_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketBookingConfirmed(
			ctx, &ticketsEntity.TicketBookingConfirmed_v1{
				// you should map v0 event to your v1 event here
				Header:        bookingConfirmedEvent.Header,
				TicketID:      bookingConfirmedEvent.TicketID,
				CustomerEmail: bookingConfirmedEvent.CustomerEmail,
				Price:         bookingConfirmedEvent.Price,
				BookingID:     bookingConfirmedEvent.BookingID,
			},
		)
	case "TicketReceiptIssued_v0":
		receiptIssuedEvent, err := unmarshalDataLakeEvent[ticketReceiptIssued_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketReceiptIssued(
			ctx, &ticketsEntity.TicketReceiptIssued_v1{
				// you should map v0 event to your v1 event here
				Header:        receiptIssuedEvent.Header,
				TicketID:      receiptIssuedEvent.TicketID,
				ReceiptNumber: receiptIssuedEvent.ReceiptNumber,
				IssuedAt:      receiptIssuedEvent.IssuedAt,
			},
		)
	case "TicketPrinted_v0":
		ticketPrintedEvent, err := unmarshalDataLakeEvent[ticketPrinted_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketPrinted(
			ctx, &ticketsEntity.TicketPrinted_v1{
				// you should map v0 event to your v1 event here
				Header:   ticketPrintedEvent.Header,
				TicketID: ticketPrintedEvent.TicketID,
				FileName: ticketPrintedEvent.FileName,
			},
		)
	case "TicketRefunded_v0":
		ticketRefundedEvent, err := unmarshalDataLakeEvent[ticketRefunded_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketRefunded(
			ctx, &ticketsEntity.TicketRefunded_v1{
				// you should map v0 event to your v1 event here
				Header:   ticketRefundedEvent.Header,
				TicketID: ticketRefundedEvent.TicketID,
			},
		)
	default:
		return fmt.Errorf("unknown event %s", event.EventName)
	}
}

func MigrateReadModel(ctx context.Context, dl ticketsDB.EventsRepository, rm ticketsDB.OpsBookingReadModel) error {
	var events []ticketsEntity.DataLakeEvent

	logger := log.FromContext(ctx)
	logger.Info("Migrating read model")

	timeout := time.Now().Add(time.Second * 10)

	// events are not immediately available in the data lake, so we need to wait for them
	for {
		var err error
		events, err = dl.GetEvents(ctx)
		if err != nil {
			return fmt.Errorf("could not get events from data lake: %w", err)
		}
		if len(events) > 0 {
			break
		}

		if time.Now().After(timeout) {
			return fmt.Errorf("timeout while waiting for events in data lake")
		}

		time.Sleep(time.Millisecond * 100)
	}

	logger.With("events_count", len(events)).Info("Has events to migrate")

	for _, event := range events {
		start := time.Now()

		logger := log.FromContext(ctx)
		logger.With(
			"event_name", event.EventName,
			"event_id", event.EventID,
		).Info("Migrating event")

		err := migrateEvent(ctx, event, rm)
		if err != nil {
			return fmt.Errorf("could not migrate event %s (%s): %w", event.EventID, event.EventName, err)
		}

		logger.With("duration", time.Since(start)).Info("Event migrated")
	}

	return nil
}
