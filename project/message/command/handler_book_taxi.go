package command

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) BookTaxi(
	ctx context.Context,
	command *ticketsEntity.BookTaxi,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Book Taxi")
	response, err := h.bookFlightService.BookTaxi(
		ctx,
		ticketsEntity.BookTaxiRequest{
			CustomerEmail:      command.CustomerEmail,
			NumberOfPassengers: command.NumberOfPassengers,
			PassengerName:      command.CustomerName,
			ReferenceId:        command.ReferenceID,
			IdempotencyKey:     command.IdempotencyKey,
		},
	)

	if err != nil {
		errPub := h.eventBus.Publish(
			ctx, ticketsEntity.TaxiBookingFailed_v1{
				Header:        ticketsEntity.NewMessageHeader(),
				ReferenceID:   command.ReferenceID,
				FailureReason: "Out of tickets",
			},
		)
		if errPub != nil {
			logger.Warn("Failed to publish taxi booking failed event")
			return errPub
		}
		return err
	}

	err = h.eventBus.Publish(
		ctx, ticketsEntity.TaxiBooked_v1{
			Header:        ticketsEntity.NewMessageHeader(),
			TaxiBookingID: response.TaxiBookingId,
			ReferenceID:   command.ReferenceID,
		},
	)
	if err != nil {
		logger.Warn("Failed to publish taxi booked event")
		return err
	}

	return nil
}
