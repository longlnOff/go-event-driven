package http

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus *cqrs.EventBus
	repo     TicketsRepository
}

type TicketsRepository interface {
	FindAll(ctx context.Context) ([]ticketsEntity.Ticket, error)
}
