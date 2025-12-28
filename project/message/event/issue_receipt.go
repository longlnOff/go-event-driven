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
	TicketReceiptIssued, err := h.receiptsService.IssueReceipt(
		ctx,
		request,
	)
	if err != nil {
		return err
	}
	err = h.eventBus.Publish(ctx, ticketsEntity.TicketReceiptIssued{
		Header:        ticketsEntity.NewMessageHeaderWithIdempotencyKey(event.Header.IdempotencyKey),
		TicketID:      event.TicketID,
		ReceiptNumber: TicketReceiptIssued.ReceiptNumber,
		IssuedAt:      TicketReceiptIssued.IssuedAt,
	})
	if err != nil {
		return err
	}

	return err
}
