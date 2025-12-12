package event

import (
	"context"
	"log/slog"

	ticketsEntity "tickets/entities"
)

func (h Handler) CancelTicket(
	ctx context.Context,
	event ticketsEntity.TicketBookingCanceled,
) error {
	slog.Info("Appending ticket to the tracker")
	err := h.spreadsheetsAPI.AppendRow(
		ctx,
		"tickets-to-refund",
		[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency},
	)

	return err
}
