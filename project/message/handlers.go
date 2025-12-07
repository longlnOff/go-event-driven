package message

import (
	"context"
	"log/slog"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)



type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, ticketID string) error
}

func NewHandler(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
	rdb redis.UniversalClient,
	watermillLogger watermill.LoggerAdapter,
) {
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

	go func() {
		messages, err := issueReceiptSub.Subscribe(context.Background(), "issue-receipt")
		if err != nil {
			panic(err)
		}

		for msg := range messages {
			err := receiptsService.IssueReceipt(msg.Context(), string(msg.Payload))
			if err != nil {
				slog.With("error", err).Error("Error issuing receipt")
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()

	go func() {
		messages, err := appendToTrackerSub.Subscribe(context.Background(), "append-to-tracker")
		if err != nil {
			panic(err)
		}

		for msg := range messages {
			err := spreadsheetsAPI.AppendRow(
				msg.Context(),
				"tickets-to-print",
				[]string{string(msg.Payload)},
			)
			if err != nil {
				slog.With("error", err).Error("Error appending to tracker")
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()
}
