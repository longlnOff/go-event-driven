package command

import (
	"context"
	"errors"
	ticketsDB "tickets/db"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) BookShowTickets(
	ctx context.Context,
	command *ticketsEntity.BookShowTickets,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Book Show Tickets")
	booking := ticketsEntity.Booking{
		BookingID:       command.BookingID.String(),
		ShowID:          command.ShowId.String(),
		NumberOfTickets: command.NumberOfTickets,
		CustomerEmail:   command.CustomerEmail,
	}

	err := h.bookingRepository.AddBooking(ctx, booking)
	if err != nil {
		if errors.As(err, &ticketsDB.ErrNoPlacesLeft) {
			errPub := h.eventBus.Publish(
				ctx, ticketsEntity.BookingFailed_v1{
					Header:        ticketsEntity.NewMessageHeader(),
					BookingID:     command.BookingID,
					FailureReason: "Out of tickets",
				},
			)
			if errPub != nil {
				logger.Warn("Failed to publish booking event")
				return errPub
			}
			return err
		}
		if errors.Is(err, ticketsDB.ErrBookingAlreadyExists) {
			// now AddBooking is called via Pub/Sub, we are taking into account at-least-once delivery
			return nil
		}
	}

	return err
}
