package entities

type IssueReceiptRequest struct {
	TicketID       string `json:"ticket_id"`
	Price          Money  `json:"price"`
	IdempotencyKey string `json:"idempotency_key"`
}
