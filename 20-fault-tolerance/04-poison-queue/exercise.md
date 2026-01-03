# Poison Queue

Previously, we discussed that {{exerciseLink "retrying failed messages" "07-errors" "02-project-temporary-errors"}} is often a good solution to handle errors.
Then we looked into {{exerciseLink "malformed messages" "07-errors" "03-project-malformed-messages"}} and how to handle them.
But how do we decide which strategy to use for which message? And do we want the message to be retried over and over until we decide what to do?

If messages are not ordered, this isn't a big issue. A spinning message won't block other messages from being processed.
However, if our messages {{exerciseLink "are ordered" "14-message-ordering" "01-ordering"}}, it will block all messages on the topic. This happens within the same ordering key or partition.

A common strategy is to use a poison queue, also known as a dead-letter queue.
This is a separate message queue (a topic) where you move messages that can't be processed.
This allows the main topic to be no longer blocked, and you can inspect the messages and decide what to do with them.

{{tip}}

Using a poison queue comes with tradeoffs.
It's a useful tool, but, depending on your needs, you may decide not to use it.
It's not a "must have" if you can use a different strategy to avoid blocked messages.

One tradeoff is that using a poison queue affects the order of the messages.
If you rely on messages being in a specific order, a poison queue might be not the best choice.

You also need some kind of tooling to manage it.
Most often, you'll need to build something custom yourself.
We'll show how to do this in the following exercises.

Remember: if your message broker doesn't send messages in order, a poison queue may not be necessary.

{{endtip}}

Watermill provides `PoisonQueue` middleware that is ready to use.
All you need to specify is the publisher and a topic where the message should be published.
When the message handler returns an error, the message is acknowledged (removed from the current topic) and published to the poison queue instead.

```go
pq, err := middleware.PoisonQueue(publisher, "poison_queue")
// ...

router.AddMiddleware(pq)
```

Alternatively, you can use the `PoisonQueueWithFilter` middleware.
It allows you to specify a function that decides if a message should be sent to the poison queue or not based on the error returned from the handler.
You can use it together with the {{exerciseLink "Permanent Error" "07-errors" "03-project-malformed-messages"}} pattern
to retry most messages and send only some to the poison queue.

```go
pq, err := middleware.PoisonQueueWithFilter(pub, "PoisonQueue", func(err error) bool {
	var permErr PermanentError
	if errors.As(err, &permErr) && permErr.IsPermanent() {
		return true
	}
	
	return false
})
```

Finally, you can chain the `PoisonQueue` middleware with the `Retry` middleware.
The message is first retried a few times, and if it still fails, it's sent to the poison queue.

```go
router.AddMiddleware(
	middleware.PoisonQueue(publisher, "poison_queue"), 
	middleware.Retry{
		// Config
	}.Middleware,
)
```

Note the somewhat counterintuitive order of the middleware.
The error-handling middleware wraps the "next" handler, so the order is reversed.
In this scenario, the `PoisonQueue` middleware handles the error returned by the `Retry` middleware.

{{tip}}

Note that you don't need to provide the same Publisher as is used for publishing the regular messages.
You can store the poisoned messages on any other compatible Pub/Sub — for example, an SQL database.

{{endtip}}

## Exercise

Exercise path: ./20-fault-tolerance/04-poison-queue/main.go

This exercise shows a system that listens for `OrderDispatched` events and saves the `TrackingLink` for each order in the storage.

1. Sometimes the `TrackingLink` is missing. We need to investigate what went wrong. We don't want to save it.
   Check if `OrderDispatched` has a `TrackingLink` field.

2. Publish messages without `TrackingLink` to the `PoisonQueue` topic. Use the provided `Publisher` (`pub`).

We recommend using [`middleware.PoisonQueueWithFilter`](https://watermill.io/docs/middlewares/#poison) to achieve this. You can also do it by hand.
Don't move the message to the poison queue if any other error occurs.
It may be a temporary error.
We want to retry it.

{{hints}}

{{hint 1}}

If you want to implement poison queue by hand, you may need to use `cqrs.OriginalMessageFromCtx(ctx)`.
This extracts the original message from the context.

Remember to ack the message after it's published to the poison queue.

{{endhint}}

{{endhints}}
