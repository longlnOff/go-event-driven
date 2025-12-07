package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	// "github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)


func main() {
	logger := watermill.NewSlogLogger(nil)

	rdb := redis.NewClient(
		&redis.Options{
			Addr: os.Getenv("REDIS_ADDR"),
		},
	)

	subscriber, _:= redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client: rdb,
		},
		logger,
	)

	msgs, _ := subscriber.Subscribe(context.Background(), "progress")
	for msg := range msgs {
		fmt.Printf("Message ID: %v - %v", msg.UUID, string(msg.Payload))
		msg.Ack()
	}

}
