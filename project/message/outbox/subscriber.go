package outbox

import (
	"github.com/ThreeDotsLabs/watermill"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/jmoiron/sqlx"
)

func NewPostgresSubscriber(
	db *sqlx.DB,
	logger watermill.LoggerAdapter,
) *watermillSQL.Subscriber {
	postgresSub, err := watermillSQL.NewSubscriber(
		db,
		watermillSQL.SubscriberConfig{
			InitializeSchema: true,
			SchemaAdapter:    watermillSQL.DefaultPostgreSQLSchema{},
			OffsetsAdapter:   watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
		},
		logger,
	)
	if err != nil {
		panic(err)
	}
	return postgresSub
}
