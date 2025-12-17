package event

import (
	"context"
	ticketsEntity "tickets/entities"
)

type Handler struct {
	spreadsheetsAPI  SpreadsheetsAPI
	receiptsService  ReceiptsService
	ticketRepository TicketsRepository
}

func NewEventHandler(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
	ticketRepository TicketsRepository,
) *Handler {
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if ticketRepository == nil {
		panic("missing ticketRepository")
	}
	return &Handler{
		spreadsheetsAPI:  spreadsheetsAPI,
		receiptsService:  receiptsService,
		ticketRepository: ticketRepository,
	}
}

type TicketsRepository interface {
	Add(ctx context.Context, ticket ticketsEntity.Ticket) error
	Remove(ctx context.Context, ticket ticketsEntity.Ticket) error
}
