# Event Headers

Apart from the event's payload, each event has some common metadata, such as the event's ID or the time it was published.
For debugging and observability purposes, it's useful to include this as a *header*.

```go
type MessageHeader struct {
	ID             string `json:"id"`
	EventName      string `json:"event_name"`
	CorrelationID  string `json:"correlation_id"`
	PublishedAt    string `json:"published_at"`
}
```

This is just an example header. You can use any fields that make sense for your use case.

It's a good convention to have all events include the header as part of the payload:

```go
type ProductCreated struct {
	Header    MessageHeader `json:"header"`
	ProductID string        `json:"product_id"`
	Name      string        `json:"name"`
}
```

Since all headers are created in a similar way, it's a good idea to have a constructor for this:

```go
func NewMessageHeader(eventName string) MessageHeader {
	return MessageHeader{
		ID:          uuid.NewString(), 
		EventName:   eventName, 
		PublishedAt: time.Now().Format(time.RFC3339),
	}
}
```

## Exercise

Exercise path: ./05-events/03-headers/main.go

**Fill in the code so that the `ProductOutOfStock` and `ProductBackInStock` events include a header.**

The event payload should look like this:

```go
{
	"header": {
		"id": "...",
		"event_name": "ProductOutOfStock",
		"occurred_at": "2023-01-01T00:00:00Z"
    },
}
```

The `event_name` should be either `ProductOutOfStock` or `ProductBackInStock`.

The `id` can be any unique identifier, for example `watermill.NewUUID()`.
