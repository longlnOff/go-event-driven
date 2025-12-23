package main

import (
	"github.com/ThreeDotsLabs/watermill"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func PublishInTx(
	message *message.Message,
	tx *sqlx.Tx,
) error {
	// TODO: your code goes here
	publisher, err := watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		watermill.NewSlogLogger(nil),
	)
	if err != nil {
		return err
	}
	err = publisher.Publish("ItemAddedToCart", message)
	if err != nil {
		return err
	}
	return nil
}
