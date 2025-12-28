package http

import (
	"context"
	"database/sql"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus          *cqrs.EventBus
	commandBus        *cqrs.CommandBus
	ticketRepository  TicketsRepository
	showRepository    ShowsRepository
	bookingRepository BookingRepository
	opsReadModel      OpsBookingReadModel
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

type OpsBookingReadModel interface {
	AllBookingsByDate(date string) ([]ticketsEntity.OpsBooking, error)
	ReservationReadModel(ctx context.Context, bookingID string) (ticketsEntity.OpsBooking, error)
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
