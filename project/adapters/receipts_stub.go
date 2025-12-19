package adapters

import (
	"context"
	"sync"
	"tickets/entities"
)

type ReceiptsServiceStub struct {
	lock           sync.Mutex
	IssuedReceipts map[string]entities.IssueReceiptRequest
}

func (s *ReceiptsServiceStub) IssueReceipt(
	ctx context.Context,
	request entities.IssueReceiptRequest,
) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.IssuedReceipts[request.IdempotencyKey] = request
	return nil
}
