package http

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus          *cqrs.EventBus
	commandBus        *cqrs.CommandBus
	ticketRepository  TicketsRepository
	showRepository    ShowsRepository
	bookingRepository BookingRepository
}

type TicketsRepository interface {
	FindAll(ctx context.Context) ([]ticketsEntity.Ticket, error)
}
type ShowsRepository interface {
	AddShow(ctx context.Context, show ticketsEntity.Show) error
	ShowByID(ctx context.Context, showID string) (ticketsEntity.Show, error)
}

type BookingRepository interface {
	AddBooking(ctx context.Context, booking ticketsEntity.Booking) error
}
