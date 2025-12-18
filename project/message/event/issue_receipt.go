package event

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) IssueReceipt(
	ctx context.Context,
	event *ticketsEntity.TicketBookingConfirmed,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Issue Receipt")

	request := ticketsEntity.IssueReceiptRequest{
		TicketID:       event.TicketID,
		Price:          event.Price,
		IdempotencyKey: event.Header.IdempotencyKey,
	}
	err := h.receiptsService.IssueReceipt(
		ctx,
		request,
	)

	return err
}
