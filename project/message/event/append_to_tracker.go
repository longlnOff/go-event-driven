package event

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) AppendToTracker(
	ctx context.Context,
	event *ticketsEntity.TicketBookingConfirmed,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Appending ticket to the tracker")

	err := h.spreadsheetsAPI.AppendRow(
		ctx,
		"tickets-to-print",
		[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency},
	)

	return err
}
