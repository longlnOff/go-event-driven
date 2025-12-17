# Using the Event Processor

The next building block is the Event Processor.
It's responsible for subscribing to a certain topic, unmarshaling events, and calling the appropriate event handlers.
Like the Event Bus, the Event Processor hides the details of choosing the right topic and marshaling events.
You just specify which event you want to handle, and the rest is handled for you.
This makes it much easier to add new handlers and keep the configuration consistent.

Setting up the Event Processor requires a bit more configuration than the Event Bus.

To create the Event Processor, use `cqrs.NewEventProcessorWithConfig`:

```go
ep, err := cqrs.NewEventProcessorWithConfig(
	router,
	cqrs.EventProcessorConfig{
		SubscriberConstructor: func (params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return sub, nil
		},
		GenerateSubscribeTopic: func (params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return params.EventName, nil
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		Logger: logger,
	},
)
```

Here's a brief explanation of the options:

- `SubscriberConstructor`: This allows you to configure the subscriber for each event handler.
  For now, we want to use the same subscriber for all handlers, so we always return the subscriber passed to the function.
- `GenerateSubscribeTopic`: This should return the same topic that was used by EventBus to publish the message.

## Exercise

Exercise path: ./09-cqrs-events/04-event-processor/main.go

Implement the following function:

```go
func RegisterEventHandlers(
	sub message.Subscriber,
	router *message.Router,
	handlers []cqrs.EventHandler,
	logger watermill.LoggerAdapter,
) error {
	// Your logic goes here
}
```

This function should:

1. Create a new `cqrs.EventProcessor`.
2. Register all event handlers in the event processor using the `AddHandlers` method.

Remember to use the same event marshaler with the same config in both the Event Bus and the Event Processor!

The correct configuration for the marshaler is:

```go
cqrs.JSONMarshaler{
	GenerateName: cqrs.StructName,
}
```

For `GenerateSubscribeTopic`, you can use:

```go
func(params cqrs.GenerateEventsTopicParams) string {
	return params.EventName
}
```

Don't forget to call the `AddHandlers` method on the event processor!

```go
err := ep.AddHandlers(handlers...)
```
