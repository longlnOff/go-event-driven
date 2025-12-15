package message

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"

	ticketsEntity "tickets/entities"
	ticketsEvent "tickets/message/event"
)

func NewRouter(
	eventHandler *ticketsEvent.Handler,
	rdb redis.UniversalClient,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)
	AddMiddleWare(router, watermillLogger)
	issueReceiptSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: ticketsEvent.IssueReceiptConsumerGroup,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	appendToTrackerConfirmedTicketSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: ticketsEvent.AppendToTrackerConsumerGroup,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	appendToRefundTicketSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: ticketsEvent.AppendToRefundTicket,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	router.AddConsumerHandler(
		"issue-receipt-handler",
		ticketsEvent.TicketBookingConfirmedTopic,
		issueReceiptSub,
		func(msg *message.Message) error {
			event := ticketsEntity.TicketBookingConfirmed{}
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return eventHandler.IssueReceipt(
				msg.Context(),
				event,
			)
		},
	)

	router.AddConsumerHandler(
		"append-to-tracker-handler",
		ticketsEvent.TicketBookingConfirmedTopic,
		appendToTrackerConfirmedTicketSub,
		func(msg *message.Message) error {
			event := ticketsEntity.TicketBookingConfirmed{}
			err = json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return eventHandler.AppendToTrackerConfirmedTicket(
				msg.Context(),
				event,
			)
		},
	)

	router.AddConsumerHandler(
		"append-to-refund-ticket",
		ticketsEvent.TicketBookingCanceledTopic,
		appendToRefundTicketSub,
		func(msg *message.Message) error {
			event := ticketsEntity.TicketBookingCanceled{}
			err = json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return eventHandler.CancelTicket(
				msg.Context(),
				event,
			)
		},
	)

	return router
}
