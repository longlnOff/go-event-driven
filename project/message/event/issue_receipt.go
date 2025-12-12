package event

import (
	"context"
	"log/slog"
	ticketsEntity "tickets/entities"
)

func (h Handler) IssueReceipt(
	ctx context.Context,
	event ticketsEntity.TicketBookingConfirmed,
) error {
	slog.Info("Issue Receipt")

	request := ticketsEntity.IssueReceiptRequest{
		TicketID: event.TicketID,
		Price:    event.Price,
	}
	err := h.receiptsService.IssueReceipt(
		ctx,
		request,
	)

	return err
}
