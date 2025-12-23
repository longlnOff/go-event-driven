package message

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	ticketsEvent "tickets/message/event"
	ticketsOutbox "tickets/message/outbox"
)

func NewRouter(
	postgresSubscriber message.Subscriber,
	publisher message.Publisher,
	eventProcessorConfig cqrs.EventProcessorConfig,
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
	)
	if err != nil {
		panic(err)
	}

	return router
}
