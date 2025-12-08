# Consumer Handlers

Some handlers just process incoming messages and don't publish any.
It's not required to return messages from the handler function: You can just return `nil`.

If you don't need any publishing at all, use the `AddConsumerHandler` method, which does the same thing with a simpler interface.
(This method used to be called `AddNoPublisherHandler`, so you can also see it in some older examples.)

```go
router.AddConsumerHandler(
	"handler_name", 
	"subscriber_topic", 
	subscriber, 
	func(msg *message.Message) error {
		return nil
	},
)
```

Like the other method, the returned `error` is used to acknowledge or negatively acknowledge the message.

## Exercise

Exercise path: ./04-router/02-no-publisher-handlers/main.go

1. Create a new Router, and add a no-publisher handler to it.
The handler should subscribe to the `temperature-fahrenheit` topic and print the incoming values in the following format:

```text
Temperature read: 100
```

2. Call `Run()` to run the Router.
