package command

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) RefundReceipts(
	ctx context.Context,
	command *ticketsEntity.RefundTicket,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Refunding receipts")
	err := h.receiptsService.RefundReceipt(
		ctx,
		*command,
	)

	err = h.paymentsService.RefundPayment(
		ctx,
		ticketsEntity.PaymentRefund{
			TicketID:       command.TicketID,
			IdempotencyKey: command.Header.IdempotencyKey,
			RefundReason:   "refund",
		},
	)
	if err != nil {
		return err
	}
	err = h.eventBus.Publish(
		ctx,
		ticketsEntity.TicketRefunded{
			Header:   ticketsEntity.NewMessageHeaderWithIdempotencyKey(command.Header.IdempotencyKey),
			TicketID: command.TicketID,
		},
	)

	return err
}
