package message

import (
	"fmt"
	ticketsDB "tickets/db"
	ticketsEntity "tickets/entities"
	ticketsCommand "tickets/message/command"
	ticketsEvent "tickets/message/event"
	ticketsOutbox "tickets/message/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewRouter(
	redisSubscriberStore message.Subscriber,
	redisSubscriber message.Subscriber,
	postgresSubscriber message.Subscriber,
	redisPublisher message.Publisher,
	eventProcessorConfig cqrs.EventProcessorConfig,
	commandProcessorConfig cqrs.CommandProcessorConfig,
	commandHandler ticketsCommand.Handler,
	opsReadModel ticketsDB.OpsBookingReadModel,
	eventHandler *ticketsEvent.Handler,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)
	AddMiddleWare(router, watermillLogger)
	ticketsOutbox.AddForwarderHandler(postgresSubscriber, redisPublisher, router, watermillLogger)
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		eventProcessorConfig,
	)
	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(
		router,
		commandProcessorConfig,
	)
	if err != nil {
		panic(err)
	}

	router.AddConsumerHandler(
		"events_splitter",
		"events",
		redisSubscriber,
		func(msg *message.Message) error {
			eventName := eventProcessorConfig.Marshaler.NameFromMessage(msg)
			if eventName == "" {
				return fmt.Errorf("cannot get event name from message")
			}
			return redisPublisher.Publish("events."+eventName, msg)
		},
	)

	router.AddConsumerHandler(
		"events_store",
		"events",
		redisSubscriberStore,
		func(msg *message.Message) error {
			var event ticketsEntity.Event
			if err := eventProcessorConfig.Marshaler.Unmarshal(msg, &event); err != nil {
				return fmt.Errorf("cannot unmarshal event: %w", err)
			}

			eventName := eventProcessorConfig.Marshaler.NameFromMessage(msg)
			if eventName == "" {
				return fmt.Errorf("cannot get event name from message")
			}

			return eventHandler.StoreEvent(msg.Context(), event, eventName, msg.Payload)
		},
	)

	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"AppendToTracker",
			eventHandler.AppendToTracker,
		),
		cqrs.NewEventHandler(
			"TicketRefundToSheet",
			eventHandler.TicketRefundToSheet,
		),
		cqrs.NewEventHandler(
			"IssueReceipt",
			eventHandler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"StoreTicket",
			eventHandler.StoreTickets,
		),
		cqrs.NewEventHandler(
			"RemoveCanceledTicket",
			eventHandler.RemoveCanceledTicket,
		),
		cqrs.NewEventHandler(
			"PrintConfirmedTicket",
			eventHandler.PrintTickets,
		),
		cqrs.NewEventHandler(
			"CallDeadNation",
			eventHandler.CallDeadNation,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnBookingMade",
			opsReadModel.OnBookingMade,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnTicketBookingConfirmed",
			opsReadModel.OnTicketBookingConfirmed,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnTicketRefunded",
			opsReadModel.OnTicketRefunded,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnTicketPrinted",
			opsReadModel.OnTicketPrinted,
		),
		cqrs.NewEventHandler(
			"ops_read_model.OnTicketReceiptIssued",
			opsReadModel.OnTicketReceiptIssued,
		),
	)
	if err != nil {
		panic(err)
	}

	err = commandProcessor.AddHandlers(
		cqrs.NewCommandHandler(
			"RefundReceipt",
			commandHandler.RefundReceipts,
		),
	)
	if err != nil {
		panic(err)
	}

	return router
}
