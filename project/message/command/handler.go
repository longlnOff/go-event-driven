package command

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type ReceiptsService interface {
	RefundReceipt(ctx context.Context, command ticketsEntity.RefundTicket) error
}

type PaymentsService interface {
	RefundPayment(ctx context.Context, refundPayment ticketsEntity.PaymentRefund) error
}

type Handler struct {
	receiptsService ReceiptsService
	paymentsService PaymentsService
	eventBus        *cqrs.EventBus
}

func NewCommandHandler(
	receiptsService ReceiptsService,
	paymentsService PaymentsService,
	eventBus *cqrs.EventBus,
) *Handler {
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if paymentsService == nil {
		panic("missing paymentsService")
	}
	if eventBus == nil {
		panic("missing eventBus")
	}
	return &Handler{
		receiptsService: receiptsService,
		paymentsService: paymentsService,
		eventBus:        eventBus,
	}
}
