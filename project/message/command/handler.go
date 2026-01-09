package command

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type ReceiptsService interface {
	RefundReceipt(ctx context.Context, command ticketsEntity.RefundTicket) error
}

type BookFlightsService interface {
	BookFlight(
		ctx context.Context,
		request ticketsEntity.BookFlightTicketRequest,
	) (ticketsEntity.BookFlightTicketResponse, error)

	BookTaxi(
		ctx context.Context,
		request ticketsEntity.BookTaxiRequest,
	) (ticketsEntity.BookTaxiResponse, error)

	CancelFlightTickets(
		ctx context.Context,
		request ticketsEntity.CancelFlightTicketsRequest,
	) error
}

type PaymentsService interface {
	RefundPayment(ctx context.Context, refundPayment ticketsEntity.PaymentRefund) error
}

type BookingRepository interface {
	AddBooking(ctx context.Context, booking ticketsEntity.Booking) error
}

type Handler struct {
	receiptsService   ReceiptsService
	paymentsService   PaymentsService
	bookFlightService BookFlightsService
	bookingRepository BookingRepository
	eventBus          *cqrs.EventBus
}

func NewCommandHandler(
	receiptsService ReceiptsService,
	paymentsService PaymentsService,
	bookFlightService BookFlightsService,
	bookingRepository BookingRepository,
	eventBus *cqrs.EventBus,
) *Handler {
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if paymentsService == nil {
		panic("missing paymentsService")
	}
	if bookFlightService == nil {
		panic("missing bookFlightService")
	}
	if eventBus == nil {
		panic("missing eventBus")
	}
	if bookingRepository == nil {
		panic("missing bookingRepository")
	}
	return &Handler{
		receiptsService:   receiptsService,
		paymentsService:   paymentsService,
		bookFlightService: bookFlightService,
		bookingRepository: bookingRepository,
		eventBus:          eventBus,
	}
}
