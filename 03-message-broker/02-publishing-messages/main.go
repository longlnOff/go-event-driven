package main

import (
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)


func main() {
	logger := watermill.NewSlogLogger(nil)

	rdb := redis.NewClient(
		&redis.Options{
			Addr: os.Getenv("REDIS_ADDR"),
		},
	)

	publisher, _ := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: rdb,
		},
		logger,
	)

	msg1 := message.NewMessage(watermill.NewUUID(), []byte("50"))
	msg2 := message.NewMessage(watermill.NewUUID(), []byte("100"))
	publisher.Publish("progress", msg1)
	publisher.Publish("progress", msg2)
}
