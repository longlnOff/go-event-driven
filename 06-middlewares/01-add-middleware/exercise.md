# Adding Middlewares

The term "middleware" describes functions that wrap the handlers with  additional logic either before or after the handler code.
This is the same concept as in HTTP middleware, but for Pub/Sub messages.

Watermill comes with some middleware that is ready to use. You can find the full list in the [documentation](https://watermill.io/docs/middlewares/).

Here's an example of middleware that times out the handler after a given time.

```go
router := message.NewDefaultRouter(logger)

router.AddMiddleware(middleware.Timeout(time.Second * 10))

router.AddConsumerHandler(
	"handler", 
	"topic", 
	handler,
)
```

You can add multiple middleware functions to the Router. They will be executed in the order they were added.

One of the most useful middleware functions is the `Recoverer` middleware.
It catches panics in the handler and turns them into errors, so the service doesn't crash.

```go
router.AddMiddleware(middleware.Recoverer)
```

It's straightforward to write your own middleware. It's just a function that takes a handler and returns a handler.

Here's custom middleware that skips the handler for events with no name in the header.

```go
func SkipNoNameEvents(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		var header Header
		err := json.Unmarshal(msg.Payload, &header)
		if err != nil {
			return nil, err
		}
		
		if header.EventName == "" {
			fmt.Println("Skipping the event due to missing name")
			return nil, nil
		}
	
		return h(msg)
	}
}
```

{{tip}}

Notice how the middleware unmarshals the payload into the `Header` struct.
This is a handy pattern that lets you access the header even if you don't know what type of event it is.
It works as long as all your headers are consistent.

{{endtip}}

### Correlation ID

A common use case for middleware is to pass the correlation ID between requests and messages.

The correlation ID is a unique string that allows you to track what happened across services and requests.
The easiest way is to append it to each log line. By searching for it in your log system, you can isolate
all logs related to the same request, regardless of how many services and API calls were involved.

Many HTTP APIs add it as a header. You can get the same effect with events by using the message metadata.

Message metadata is a key-value store that can be used to pass additional information along with a message.
Most message brokers support metadata. Message metadata is very similar to HTTP headers, so you can use it in a similar way.
One major advantage is that you can retrieve metadata values without unmarshalling the message payload.

Watermill provides the middleware and some helper functions that should cover most use cases.
The middleware checks for any correlation ID in the incoming message's metadata and adds it to any new messages created by the handler.

```go
router.AddMiddleware(middleware.CorrelationID)
```

To add the correlation ID to a message's metadata:

```go
msg := message.NewMessage(watermill.NewUUID(), payload)
middleware.SetCorrelationID(correlationID, msg)
```

And to retrieve it:

```go
correlationID := middleware.MessageCorrelationID(msg)
```

## Exercise

Exercise path: ./06-middlewares/01-add-middleware/main.go

The provided code for this exercise is a part of a game engine backend.

The entry point is the `POST /players` HTTP endpoint, which stores the player in an in-memory database and publishes a `PlayerJoined` event.
In reaction to it, a message handler creates a team (if the player chooses a team that doesn't exist), and publishes the `TeamCreated` event.
Finally, another message handler calls an external scoreboard service in reaction to a new team being created.

```mermaid
graph LR
A[POST /players] -- PlayerJoined --> B[OnPlayerJoined]
A -- Store player --> C[In-memory Database]
B -- TeamCreated --> D[OnTeamCreated]
D -- Create Scoreboard --> F[Scoreboard Service]
```

The code is missing a correlation ID, so it's hard to debug any issues.

Your task is to propagate the correlation ID from the HTTP request through all messages' metadata, 
up to the scoreboard HTTP request:

1. Add the correlation middleware to the Router.
2. Set the correlation ID for the `PlayerJoined` event based on the incoming correlation ID from the HTTP header.
3. If the incoming correlation ID is empty, generate a new correlation ID; for example, a UUID.
4. Pass the correlation ID from the `TeamCreated` event to the scoreboard request.

To make your task easier, we left `TODO` in places where you need to add the code.
