package command

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewCommandBus(
	pub message.Publisher,
	logger watermill.LoggerAdapter,
) *cqrs.CommandBus {
	eventBus, err := cqrs.NewCommandBusWithConfig(
		pub,
		cqrs.CommandBusConfig{
			GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
				return "commands." + params.CommandName, nil
			},
			Marshaler: jsonMarshaler,
			Logger:    logger,
		},
	)
	if err != nil {
		panic(err)
	}
	return eventBus
}
