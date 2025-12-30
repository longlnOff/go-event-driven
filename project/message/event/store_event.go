package event

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) StoreEvent(
	ctx context.Context,
	event ticketsEntity.Event,
	eventName string,
	payload []byte,
) error {
	logger := log.FromContext(ctx)
	logger.Info("store event")

	err := h.eventRepository.SaveEvent(ctx, event, eventName, payload)
	if err != nil {
		logger.Error("failed to store event")
		return err
	}

	return nil
}
