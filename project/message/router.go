package message

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)



type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, ticketID string) error
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
		func (msg *message.Message) error {
			err := receiptsService.IssueReceipt(msg.Context(), string(msg.Payload))
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
		func (msg *message.Message) error {
			err := spreadsheetsAPI.AppendRow(
				msg.Context(),
				"tickets-to-print",
				[]string{string(msg.Payload)},
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
