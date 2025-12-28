package entities

import "time"

type IssueReceiptRequest struct {
	TicketID       string `json:"ticket_id"`
	Price          Money  `json:"price"`
	IdempotencyKey string `json:"idempotency_key"`
}

type IssueReceiptResponse struct {
	ReceiptNumber string    `json:"receipt_number"`
	IssuedAt      time.Time `json:"issued_at"`
}
