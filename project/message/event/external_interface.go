package event

import (
	"context"
	ticketsEntity "tickets/entities"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request ticketsEntity.IssueReceiptRequest) error
}

type FilesService interface {
	PrintTickets(ctx context.Context, ticketFile string, body string) error
}
