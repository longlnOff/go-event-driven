package event

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) TicketRefundToSheet(
	ctx context.Context,
	event *ticketsEntity.TicketBookingCanceled,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Adding ticket refund to sheet")
	err := h.spreadsheetsAPI.AppendRow(
		ctx,
		"tickets-to-refund",
		[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency},
	)

	return err
}
