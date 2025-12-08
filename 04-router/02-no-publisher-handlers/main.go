package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := watermill.NewSlogLogger(nil)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	router := message.NewDefaultRouter(logger)

	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	router.AddConsumerHandler(
		"print-f-value",
		"temperature-fahrenheit",
		sub,
		func (msg *message.Message) error {
			fmt.Println("Temperature read:", string(msg.Payload))
			return nil
		},
	)

	router.Run(context.Background())
}
