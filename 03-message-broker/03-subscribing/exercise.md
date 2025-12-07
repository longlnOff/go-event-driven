# Subscribing for messages

To receive messages, you need to *subscribe* to the topic they're being published on.

```go
type Subscriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan *Message, error)
	Close() error
}
```

Creating a subscriber is very similar to creating a publisher:

```go
subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
	Client: rdb,
}, logger)
```

The idea is very similar to the publishing side: You must specify a topic, so the Pub/Sub knows which messages to deliver to you.
Most of the time, this is the same string you used for publishing.

```go
messages, err := subscriber.Subscribe(context.Background(), "orders")
if err != nil {
	panic(err)
}
```

`Subscribe` returns a channel of messages.
Any new message published on the topic will be delivered to it.

You can use it like any Go channel.
Usually, you'll want to use a `for` loop to process incoming messages.

Most Pub/Subs deliver a single message at a time.
You need to let the broker know that a message has been correctly processed with a *message acknowledgement* (*ack* in short).
Once you do it, the broker removes the message from the topic, and delivers the next one.

Watermill's messages expose the `Ack()` method, which does this.
The correct iteration looks like this:

```go
for msg := range messages {
	orderID := string(msg.Payload)
	fmt.Println("New order placed with ID:", orderID)
	msg.Ack()
}
```

**It's easy to miss this step, but it's crucial.**
If you notice that your subscriber receives a single message and then stops,
**it's probably because you forgot to `Ack()` the message.**

### Closing the Subscriber

While each `Publish` is a one-time operation, `Subscribe` starts an asynchronous worker process.

To close it, you can either call the `Close()` method on the subscriber, or cancel the context passed to `Subscribe`.

{{tip}}

### The Context Primer

If you're not familiar with `context.Context`, here's a short introduction.

The Context is used mainly for two purposes:

* Canceling long-running operations, either via timeouts or explicit cancellation.
* Passing arbitrary values between functions.

The base context is created with `context.Background()`.
It's an empty context, that has no behavior.

```go
ctx := context.Background()
```

You can create a new context from an existing one with `context.WithCancel()`.
Calling the `cancel` function will cancel the context and all contexts created from it.

```go
ctx, cancel := context.WithCancel(context.Background())
```

You can also create a context with a timeout, which will cancel itself after the specified duration.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
```

You will see `context.Context` as the first argument of many functions.
It usually indicates that the function does an external request or a long-running operation.

{{endtip}}

## Exercise

Exercise path: ./03-message-broker/03-subscribing/main.go

Create a Redis Stream subscriber and subscribe to the `progress` topic.

Print the incoming messages in the following format:

```text
Message ID: a16b6ab0-8c29-48f5-9d26-b508906af976 - 50
Message ID: d33c5cac-1cce-4783-b931-7ecba25fa7dc - 100
```

Don't forget to *ack* the messages.

For now, use `context.Background()` where a context is needed.
