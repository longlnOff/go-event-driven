package main

import (
	"context"
	"sync"
)

type IssueReceiptRequest struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request IssueReceiptRequest) error
}

type ReceiptsServiceStub struct {
	lock           sync.Mutex
	IssuedReceipts []IssueReceiptRequest
}

func (s *ReceiptsServiceStub) IssueReceipt(
	ctx context.Context,
	request IssueReceiptRequest,
) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.IssuedReceipts = append(s.IssuedReceipts, request)

	return nil
}
