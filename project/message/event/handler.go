package event

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request ticketsEntity.IssueReceiptRequest) (ticketsEntity.IssueReceiptResponse, error)
}

type FilesService interface {
	UpLoadFile(ctx context.Context, ticketFile string, body string) error
}

type DeadNationService interface {
	CallDeadNation(ctx context.Context, booking ticketsEntity.DeadNationBooking) error
}

type TicketsRepository interface {
	Add(ctx context.Context, ticket ticketsEntity.Ticket) error
	Remove(ctx context.Context, ticket ticketsEntity.Ticket) error
}

type ShowsRepository interface {
	AddShow(ctx context.Context, show ticketsEntity.Show) error
	ShowByID(ctx context.Context, showID string) (ticketsEntity.Show, error)
}

type Handler struct {
	spreadsheetsAPI   SpreadsheetsAPI
	receiptsService   ReceiptsService
	fileService       FilesService
	deadNationService DeadNationService
	ticketRepository  TicketsRepository
	showRepository    ShowsRepository
	eventBus          *cqrs.EventBus
}

func NewEventHandler(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
	fileService FilesService,
	deadNationService DeadNationService,
	ticketRepository TicketsRepository,
	showRepository ShowsRepository,
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
	if deadNationService == nil {
		panic("missing deadNationService")
	}
	if ticketRepository == nil {
		panic("missing ticketRepository")
	}
	if showRepository == nil {
		panic("missing showRepository")
	}
	if eventBus == nil {
		panic("missing eventBus")
	}
	return &Handler{
		spreadsheetsAPI:   spreadsheetsAPI,
		receiptsService:   receiptsService,
		fileService:       fileService,
		deadNationService: deadNationService,
		ticketRepository:  ticketRepository,
		showRepository:    showRepository,
		eventBus:          eventBus,
	}
}
