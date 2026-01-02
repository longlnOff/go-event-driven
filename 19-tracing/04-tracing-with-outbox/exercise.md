## Tracing Pub/Sub Outbox

It's important to not miss traces when you use the outbox pattern.
This may happen if you forget to decorate your outbox publisher.

To ensure you have the proper `traceparent` metadata value, you need to decorate the publisher that publishes the message to the outbox.

```go
func PublishInTx(
	msg *message.Message,
	tx *sql.Tx,
	logger watermill.LoggerAdapter,
) error {
	var publisher message.Publisher
	var err error

	// this publisher publishes "enveloped" message, 
	
	publisher, err = watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox publisher: %w", err)
	}

	// TODO: It's worth decorating this one as well to see the forwarding operation in the trace.
	// (Forwarding is done based on the enveloped message.)

	// This publisher publishes "our" message.
	publisher = forwarder.NewPublisher(publisher, forwarder.PublisherConfig{
		ForwarderTopic: outboxTopic,
	})
	
	// TODO: You should decorate your publisher here.

	return publisher.Publish("ItemAddedToCart", msg)
}
```

## Exercise

Exercise path: ./19-tracing/04-tracing-with-outbox/main.go

In this exercise, we will work again on the {{exerciseLink "outbox exercise" "11-outbox" "08-publishing-events-with-forwarder"}}.
You can refer to that exercise to refresh your memory.

**Use the publisher decorator `TracingPublisherDecorator` in the proper place, so you don't miss the trace context within the outbox.**

{{tip}}

In the exercise, the bare minimum to implement is decorating the "last" publisher in that function.

We recommend also decorating the SQL publisher, so you can trace the forwarding operation.
As a reminder: During forwarding, we get the destination topic from the enveloped message.
We have two messages and want to add the `traceparent` metadata to both of them.
In other words, you need to decorate both `watermillSQL.NewPublisher` and `forwarder.NewPublisher`.

To refresh how this works, we recommend going back to the {{exerciseLink "outbox module" "11-outbox" "07-forwarding-with-outbox"}}.

{{endtip}}
