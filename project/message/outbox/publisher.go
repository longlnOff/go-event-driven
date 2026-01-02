package outbox

import (
	"context"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TracePublisherDecorator struct {
	message.Publisher
}

func (c TracePublisherDecorator) Publish(topic string, messages ...*message.Message) error {
	for i := range messages {
		otel.GetTextMapPropagator().Inject(messages[i].Context(), propagation.MapCarrier(messages[i].Metadata))
	}
	return c.Publisher.Publish(topic, messages...)
}

func NewPublisherForDb(
	ctx context.Context,
	tx *sqlx.Tx,
) (message.Publisher, error) {
	var publisher message.Publisher

	publisher, err := watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		watermill.NewSlogLogger(log.FromContext(ctx)),
	)
	if err != nil {
		return nil, err
	}

	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}
	publisher = TracePublisherDecorator{publisher}

	publisher = forwarder.NewPublisher(
		publisher,
		forwarder.PublisherConfig{
			ForwarderTopic: outboxTopic,
		},
	)
	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}
	publisher = TracePublisherDecorator{publisher}

	return publisher, nil
}
