package command

import (
	"context"
	"errors"
	"fmt"
	ticketsDB "tickets/db"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) BookFlight(
	ctx context.Context,
	command *ticketsEntity.BookFlight,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Book Flight For Tickets")
	bookFlightRequest := ticketsEntity.BookFlightTicketRequest{
		CustomerEmail:  command.CustomerEmail,
		FlightID:       command.FlightID,
		PassengerNames: command.Passengers,
		ReferenceId:    command.ReferenceID,
		IdempotencyKey: command.IdempotencyKey,
	}

	bookFlightResponse, err := h.bookFlightService.BookFlight(ctx, bookFlightRequest)
	if err != nil {
		if errors.As(err, &ticketsDB.ErrNoPlacesLeft) {
			errPub := h.eventBus.Publish(
				ctx, ticketsEntity.FlightBookingFailed_v1{
					Header:        ticketsEntity.NewMessageHeader(),
					FlightID:      command.FlightID,
					ReferenceID:   command.ReferenceID,
					FailureReason: "Out of flight tickets",
				},
			)
			if errPub != nil {
				logger.Warn("Failed to publish booking event")
				return errPub
			}
			return err
		}
	}
	err = h.eventBus.Publish(
		ctx, ticketsEntity.FlightBooked_v1{
			Header:      ticketsEntity.NewMessageHeader(),
			FlightID:    command.FlightID,
			TicketIDs:   bookFlightResponse.TicketIds,
			ReferenceID: command.ReferenceID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish FlightBooked_v1 event: %w", err)
	}

	return nil
}
