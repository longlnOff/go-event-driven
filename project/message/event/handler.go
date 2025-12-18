package event

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	spreadsheetsAPI  SpreadsheetsAPI
	receiptsService  ReceiptsService
	fileService      FilesService
	ticketRepository TicketsRepository
	eventBus         *cqrs.EventBus
}

func NewEventHandler(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
	fileService FilesService,
	ticketRepository TicketsRepository,
	eventBus *cqrs.EventBus,
) *Handler {
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if fileService == nil {
		panic("missing fileService")
	}
	if ticketRepository == nil {
		panic("missing ticketRepository")
	}
	if eventBus == nil {
		panic("missing eventBus")
	}
	return &Handler{
		spreadsheetsAPI:  spreadsheetsAPI,
		receiptsService:  receiptsService,
		fileService:      fileService,
		ticketRepository: ticketRepository,
		eventBus:         eventBus,
	}
}

type TicketsRepository interface {
	Add(ctx context.Context, ticket ticketsEntity.Ticket) error
	Remove(ctx context.Context, ticket ticketsEntity.Ticket) error
}
