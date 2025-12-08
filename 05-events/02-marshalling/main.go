package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

type PaymentCompleted struct {
	PaymentID   string `json:"payment_id"`
	OrderID     string `json:"order_id"`
	CompletedAt string `json:"completed_at"`
}

type PaymentConfirmed struct {
	OrderID     string `json:"order_id"`
	ConfirmedAt string `json:"confirmed_at"`
}

func main() {
	logger := watermill.NewSlogLogger(nil)

	router := message.NewDefaultRouter(logger)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}


	router.AddHandler(
		"confirmed-payment-handler",
		"payment-completed",
		sub,
		"order-confirmed",
		pub,
		func (msg *message.Message)  ([]*message.Message, error) {
			data:= PaymentCompleted{}
			err := json.Unmarshal(msg.Payload, &data)
			if err != nil {
				return []*message.Message{},  err
			}
			event := PaymentConfirmed{
				OrderID: data.OrderID,
				ConfirmedAt: data.CompletedAt,
			}
			payload, err := json.Marshal(event)
			if err != nil {
				return []*message.Message{},  err
			}
			
			m := message.NewMessage(watermill.NewUUID(), payload)		
			return []*message.Message{m}, nil	
		},
	)
	

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
