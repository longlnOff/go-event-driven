package message

import (
	ticketsCommand "tickets/message/command"
	ticketsEvent "tickets/message/event"
	ticketsOutbox "tickets/message/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewRouter(
	postgresSubscriber message.Subscriber,
	publisher message.Publisher,
	eventProcessorConfig cqrs.EventProcessorConfig,
	commandProcessorConfig cqrs.CommandProcessorConfig,
	commandHandler ticketsCommand.Handler,
	eventHandler *ticketsEvent.Handler,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)
	AddMiddleWare(router, watermillLogger)
	ticketsOutbox.AddForwarderHandler(postgresSubscriber, publisher, router, watermillLogger)
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
