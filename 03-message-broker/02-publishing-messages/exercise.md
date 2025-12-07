# Publishing Messages

In the next couple of modules, we'll start with the basics of message-based systems. 
We'll make sure everyone is on the same page and has a solid foundation for the more advanced parts ahead.

### Picking the library

When you serve a website, you don't handle the HTTP protocol manually.
You use a library that does the heavy lifting for you and gives you a nice API to define all the endpoints.

Most popular Pub/Subs have their own SDK in Go.
But they are usually low-level libraries, and working with each Pub/Sub is quite different.
They also don't support features like middlewares or message marshalling out of the box.

Back in 2018, we thought working with messages should be as easy as working with HTTP requests.
There was no library in Go that allowed it, so we decided to create it.
It's called [Watermill](https://watermill.io/), and we've been using it in production for many different projects since then.

**We'll be using Watermill for the rest of this training.
It will help you focus on the high-level concepts instead of the low-level details.**
Our goal is to teach you the principles behind event-driven architecture, not the specifics of a particular Pub/Sub.

You can find the documentation for Watermill at [watermill.io](https://watermill.io/).

We designed Watermill as a lightweight library, not a framework, so there's no vendor lock-in.
If you prefer to use anything else, it should be straightforward to translate the examples to other languages and libraries.
**You can apply the event-driven principles in any other programming language or library.**

### The Publisher

Watermill hides all the complexity of Pub/Subs behind just two interfaces: the `Publisher` and the `Subscriber`.

For now, let's consider the first one.

```go
type Publisher interface {
	Publish(topic string, messages ...*Message) error
	Close() error
}
```

To publish a message, you need to pass a `topic` and a slice of messages.

To *publish* a message means to append it to the given topic.
Anyone who subscribes to the same topic will receive the messages on it in a first-in, first-out (FIFO) fashion.

{{tip}}

FIFO is a common way to deliver messages, but it can vary depending on the Pub/Sub and how it's configured.

This is true for many behaviors we describe in this training.
Always check the documentation of the Pub/Sub you use to confirm it works as you expect.

{{endtip}}

How you create the publisher instance depends on the Pub/Sub you choose.
Each library provides its own constructors.

Here's how to create one for Redis Streams:

```go
logger := watermill.NewSlogLogger(nil)

rdb := redis.NewClient(&redis.Options{
	Addr: os.Getenv("REDIS_ADDR"),
})

publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
	Client: rdb,
}, logger)
```

{{tip}}

The `watermill.NewSlogLogger`'s argument is optional slog instance (when nil, it uses [`slog.Default()`](https://pkg.go.dev/log/slog#Default)).

{{endtip}}

You need the following imports to make the code above work:

```go
import (
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)
```

To create a message, use the `NewMessage` constructor.
It takes just two arguments.

```go
msg := message.NewMessage(watermill.NewUUID(), []byte(orderID))
```

The first argument is the message's UUID, which is used mainly for debugging.
Most of the time, any kind of UUID is fine.
Watermill provides helper functions to generate them.

The second argument is the payload. 
Like in HTTP request's body, it's a slice of bytes, so it can be anything you want, as long as you can marshal it.
You can send a string, a JSON object, or even a binary file.

To publish the message, call the `Publish` method on the publisher:

```go
err := publisher.Publish("orders", msg)
```

Remember to handle the error, as publishing works over the network and can fail for many reasons.

## Exercise

Exercise path: ./03-message-broker/02-publishing-messages/main.go

1. **Create a Redis Streams publisher.**

2. **Publish two messages on the `progress` topic.**
The first one's payload should be `50`, and the second one's should be `100`.

To get the necessary dependencies, either run `go get` for them individually or run `go mod tidy` after adding the imports.
