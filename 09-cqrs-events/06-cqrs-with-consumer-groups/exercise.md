# CQRS with Consumer Groups

Remember how we used Consumer Groups to let multiple consumers process the same messages in {{exerciseLink "the previous exercise" "03-message-broker" "05-consumer-groups"}}?
Now, we need to apply this idea to the event processor.

In `NewEventHandler`, you can specify a handler name. Each handler name must be unique, so it’s useful for generating a consumer group name.

Recall the `SubscriberConstructor` option in the event processor config.
This option is usually used to create a unique consumer group for each handler.
`SubscriberConstructor` is called for every handler, so you can create a separate subscriber with a unique consumer group name for each event handler.


```go
return cqrs.NewEventProcessorWithConfig(
	cqrs.EventProcessorConfig{
		// ... 
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return redisstream.NewSubscriber(redisstream.SubscriberConfig{
				Client:        rdb, 
				ConsumerGroup: "svc-tickets." + params.HandlerName,
			}, logger)
		},
		// ...
    },
)
```

Consumer groups were explained in depth in {{exerciseLink "the previous exercise" "03-message-broker" "05-consumer-groups"}}.
If you need a refresher, check it out.

{{tip}}

Convenience usually comes with a cost.
When using handler names to generate consumer group names, changing a handler name accidentally changes the consumer group name.

This creates a new consumer group, which may cause you to miss some messages that haven't been processed yet.

Simply "being careful" is not a reliable strategy in production.
We'll cover better ways to handle this in the alerting module.

{{endtip}}

## Exercise

Exercise path: ./09-cqrs-events/06-cqrs-with-consumer-groups/main.go

Implement the `NewEventProcessor` function that returns a `*cqrs.EventProcessor` with a RedisStream subscriber using a consumer group.

The expected signature is:

```go
func NewEventProcessor(
	rdb *redis.Client, 
	router *message.Router,
    marshaler cqrs.CommandEventMarshaler,
	logger watermill.LoggerAdapter,
) (*cqrs.EventProcessor, error) {
```

Use the following for `GenerateSubscribeTopic`: 

```go
func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
	return params.EventName, nil
}
```

The consumer group name doesn't matter. The only requirement is that it should be different for each event handler. 

A common approach is to include the service name in it to avoid handler name conflicts across services.
For example, `"svc-users." + handlerName`.
