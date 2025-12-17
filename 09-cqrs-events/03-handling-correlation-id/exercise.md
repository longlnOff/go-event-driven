# Handling the Correlation ID

In the previous implementation, we were {{exerciseLink "manually setting" "06-middlewares" "03-project-correlation-id"}} the `"correlation_id"` metadata in our message header.

There are at least two ways to handle this within the event bus:
- We can modify the marshaler.
- We can implement a publisher decorator.

### Publisher Decorator

The decorator pattern isn't very common in Go, but it can be very useful. It lets you change an object's behavior without modifying its code.

You can think of it like middleware, but for interface implementations.

An example of the decorator implementation is as follows:

```go
type Sender interface {
	Send(message string) error
}

type QuotingDecorator struct {
	Sender
}

func (q QuotingDecorator) Send(message string) error {
	message = fmt.Sprintf(`"%s"`, message)

	return q.Sender.Send(message)
}
```

{{tip}}

Sometimes, you may want to decorate an object that isn't constructed as an interface.
For example, `redisstream.NewPublisher` returns a `*redisstream.Publisher` type, which is a concrete type, so you can't simply do:


```go
pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
    Client: rdb,
}, logger)

pub = SomeDecorator{pub}
```

This is because `pub` has a concrete type and cannot be assigned to `SomeDecorator`.

You need to declare a variable of the interface type that you want to decorate (for example, `message.Publisher`) and assign the decorated object to it:

```go
var pub message.Publisher

pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
    Client: rdb,
}, logger)

pub = SomeDecorator{pub}
```

{{endtip}}

The decorator uses the same interface as the object it wraps.
It can modify arguments, return values, or even call the decorated object multiple times.

### Further Reading

If you want to learn more about the use cases of the decorator pattern, 
you should check out the [Increasing Cohesion in Go with Generic Decorators](https://threedots.tech/post/increasing-cohesion-in-go-with-generic-decorators/) article on our blog!

## Exercise

Exercise path: ./09-cqrs-events/03-handling-correlation-id/main.go

Let's implement a decorator for the publisher. 
This decorator should add the correlation ID to the message metadata key `correlation_id`. 
The decorator should be called `CorrelationPublisherDecorator`.

You can retrieve the correlation ID from the context using the `CorrelationIDFromContext` function.

Remember to set the metadata for each incoming message! Each message has its own context and metadata.
