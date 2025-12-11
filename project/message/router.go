package message

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"

	ticketsEntity "tickets/entities"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request ticketsEntity.IssueReceiptRequest) error
}

func NewRouter(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
	rdb redis.UniversalClient,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)
	issueReceiptSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "issue-receipt",
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	appendToTrackerSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "append-to-tracker",
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	router.AddConsumerHandler(
		"issue-receipt-handler",
		"issue-receipt",
		issueReceiptSub,
		func(msg *message.Message) error {
			payload := ticketsEntity.IssueReceiptPayload{}
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}
			request := ticketsEntity.IssueReceiptRequest{
				TicketID: payload.TicketID,
				Price:    payload.Price,
			}
			err = receiptsService.IssueReceipt(msg.Context(), request)
			if err != nil {
				slog.With("error", err).Error("Error issuing receipt")
				return err
			}

			return nil
		},
	)

	router.AddConsumerHandler(
		"append-to-tracker-handler",
		"append-to-tracker",
		appendToTrackerSub,
		func(msg *message.Message) error {
			payload := ticketsEntity.AppendToTrackerPayload{}
			err = json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}
			err := spreadsheetsAPI.AppendRow(
				msg.Context(),
				"tickets-to-print",
				[]string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency},
			)
			if err != nil {
				slog.With("error", err).Error("Error appending to tracker")
				return err
			}
			return nil
		},
	)

	return router
}
