# Watermill Router: Handlers

The `Publisher` and `Subscriber` abstract away many details of the underlying Pub/Sub, but they are still quite low-level interfaces. 
Watermill's high-level API is the `Router`.

The `Router` is similar to an HTTP router and gives you a convenient way to define message handlers.

Creating a router is as simple as:

```go
router := message.NewDefaultRouter(logger)
```

You can define a new handler with the `AddHandler` method. A message handler is a function that
receives a message from a given topic and publishes another message (or messages) to another topic.

You need to pass a publisher and subscriber along with some other parameters:

```go
router.AddHandler(
	"handler_name", 
	"subscriber_topic", 
	subscriber, 
	"publisher_topic", 
	publisher, 
	func(msg *message.Message) ([]*message.Message, error) {
		newMsg := message.NewMessage(watermill.NewUUID(), []byte("response"))
		return []*message.Message{newMsg}, nil
	},
)
```

The handler name is used mostly for debugging.
It can be any string you want, but it needs to be unique across handlers within the same Router.

The Router handles the publisher and subscriber orchestration for you.
You just define the input and output topics and the handler function with the message processing logic.

All messages that the handler function returns will be published to the given topic.

The returned `error` has a key role.
If the handler function returns `nil`, the message is acknowledged.
If the handler returns an error, a negative acknowledgement is sent.
You don't need to worry about calling `Ack()` or `Nack()` anymore.
Just make sure to return the proper error.

After adding all the handlers, you need to start the Router:

```go
err = router.Run(context.Background())
```

`Run` is blocking. If you don't run it in a separate goroutine, make sure you add all your handlers first.

## Exercise

Exercise path: ./04-router/01-handlers/main.go

**Add a new handler to the Router.**

It should subscribe to values from the `temperature-celsius` topic and publish the converted values to the `temperature-fahrenheit` topic.
You can use the included `celsiusToFahrenheit` function to convert the values.

{{tip}}

You don't need to call `Ack()` or `Nack()` in the Router's handler function.
The message is acknowledged if the returned error is `nil`.

{{endtip}}
