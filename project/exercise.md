# Project: Migrate to Events

Perhaps you noticed that the current messages in the project aren't events.
It's not a problem, though. We can now migrate to events and get the benefits of using them.

In the current form, our webhook handler knows much of what happens outside of it.
It publishes a message for each action that needs to happen (issuing a receipt or appending the ticket to the tracker).
This is not a big deal right now, but handlers like this tend to grow and become hard to change.
We also lose many benefits that using events brings.

We can easily improve this by replacing both messages with a proper event.

Let's recap {{exerciseLink "the beginning" "05-events" "01-events"}} of the module.
Here's what you need to keep in mind about events:

- They're facts: They describe something that already happened.
- They're immutable: Once published, they can't be changed.
- They should be expressed as verbs in past tense, like `UserSignedUp`, `OrderPlaced`, or `AlarmTriggered`.

## Exercise

Exercise path: ./project

1. **Publish the event.**

Instead of two messages published on the `issue-receipt` and `append-ticket` topics,
make the HTTP handler **publish a single `TicketBookingConfirmed` event on the `TicketBookingConfirmed` topic.**

The event should have the following form:

```go
type TicketBookingConfirmed struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}
```

The header can look like this:

```go
type MessageHeader struct {
	ID          string    `json:"id"`
	PublishedAt time.Time `json:"published_at"`
}

func NewMessageHeader() MessageHeader {
	return MessageHeader{
		ID:          uuid.NewString(),
		PublishedAt: time.Now().UTC(),
	}
}
```

2. **Update the Router handlers to subscribe to the `TicketBookingConfirmed` topic.**

Unmarshal the event payload into the `TicketBookingConfirmed` struct.
It has all required details to issue a receipt and append the ticket to the tracker.

{{tip}}

Important: **You need two subscribers, each with a unique consumer group, just like in the {{exerciseLink "consumer groups exercise" "03-message-broker" "05-consumer-groups"}}.**
Otherwise, only one handler would receive each event.

{{endtip}}

{{tip}}

We'll add more event-based features in the upcoming modules.
This might be a good moment to refactor and move event handlers into their own files (similar to how the HTTP router is organized in the example solution).

It could look like this:

```go
// message/event/handler.go

type Handler struct {
	spreadsheetsAPI SpreadsheetsAPI
	receiptsService ReceiptsService
}

// message/event/append_to_tracker.go

func (h Handler) AppendToTracker(ctx context.Context, event entities.TicketBookingConfirmed) error {
    slog.Info("Appending ticket to the tracker")
    // ...
}

// message/router.go

router.AddConsumerHandler(
		"append_to_tracker",
		"TicketBookingConfirmed",
		appendToTrackerSub,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return handler.AppendToTracker(msg.Context(), event)
		},
	)
```

It's optional, but it can help you keep the code organized as the project grows.

{{endtip}}
