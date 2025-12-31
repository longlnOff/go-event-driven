package event

import (
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewEventBus(
	pub message.Publisher,
	logger watermill.LoggerAdapter,
) *cqrs.EventBus {
	eventBus, err := cqrs.NewEventBusWithConfig(
		pub,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				event, ok := params.Event.(entities.Event)
				if !ok {
					return "", fmt.Errorf("invalid event type: %T doesn't implement entities.Event", params.Event)
				}

				if event.IsInternal() {
					// Publish directly to the per-event topic
					return "internal-events.svc-tickets." + params.EventName, nil
				} else {
					// Publish to the "events" topic, so it will be stored to the data lake and forwarded to the
					// per-event topic
					return "events", nil
				}
			},
			Logger:    logger,
			Marshaler: jsonMarshaler,
		},
	)
	if err != nil {
		panic(err)
	}
	return eventBus
}
