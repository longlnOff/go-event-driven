# Internal Events

You may be afraid to publish any new events after reading so much about backward compatibility.
However, there is one strategy that allows us to publish events without the need for any backward compatibility guarantees:
We can introduce internal events to our system.

Internal events are events consumed only by one service or team.
They should not be published to the data lake (if you use one).
They don't need backward compatibility guarantees. It's easier to change them in the future.
Thanks to that, you can change them without fear of breaking other services.
Think of them as encapsulation (like public/private methods).

There are multiple ways of naming them; in most sources you can find the following:
- For internal: private
- For external: public, integration

## When to Use Internal Events

There is no simple answer here.
On the one hand, if you are using internal events, you are making your external contract smaller.
On the other hand, you are exposing less information about your system to other teams for integration or data analytics.
**It's a tradeoff, and you need to decide if this event may change and whether it may be useful for other teams.**

Internal events may be also a good choice for a very technical events that are not related to the domain.

If you are not sure, you can always start with an internal event and expose it later.

There is no single way of implementing internal events.
We will show you how we do it in our projects.

We'll create a separate topic prefix for internal events: `internal-events.svc-tickets.`.
This clearly indicates that this topic is used only for internal events of the `svc-tickets`.
Nobody else should consume it.

A good use case for internal events might be `InternalOpsReadModelUpdated`.
We may emit it after each update of the read model and use it for sending SSE updates to the frontend.
Nobody else should depend on that event.
We may change this event at any point, without breaking any contract.

{{tip}}

Server-sent events (SSE) are out of the scope of this training, but Watermill has support for them out of the box.
You can read more in the [Watermill SSE example](https://github.com/ThreeDotsLabs/watermill/tree/master/_examples/real-world-examples/server-sent-events)
and in the [Live website updates with Go, SSE, and htmx](https://threedots.tech/post/live-website-updates-go-sse-htmx/) blog post.

{{endtip}}

Our event can be as simple as this:

```go
type InternalOpsReadModelUpdated struct {
	Header MessageHeader `json:"header"`
	
	BookingID uuid.UUID `json:"booking_id"`
}
```

It should be emitted after each update of the read model. If we used SSE, it would send the update of the content to the frontend.

**We don't need to use outbox here.** It's not a disaster if this event is lost since we'll use it for SSE.
It will also be less expensive in terms of resources. We can emit it directly to Redis Streams.

It would be good to have an explicit way to know if an event is internal or not.
Checking the prefix of the struct name doesn't sound explicit enough.
Instead, let's define the `Event` interface:

```go
type Event interface {
	IsInternal() bool
}
```

`InternalOpsReadModelUpdated` should return `true`:

```go
func (i InternalOpsReadModelUpdated) IsInternal() bool {
	return true
}
```

It should return `false` for all non-internal events.

## Data Lake

(If you skipped the data lake module, you can ignore this part.)

We want to ensure that this event won't be sent to the data lake.
To do that, we can change the logic of the event bus so that it will publish internal events directly to the per-event topic.
For `InternalOpsReadModelUpdated`, it will be `internal-events.svc-tickets.InternalOpsReadModelUpdated`.

```mermaid
graph LR
    A[Event Bus] --> B['events' topic]
    B --> C[Data lake consumer]
    B --> D[Events forwarder]
    C --> E[Data lake]
    D --> F['events.BookingMade' topic]
    D --> G['events.TicketBookingConfirmed' topic]
    D --> H['events.TicketReceiptIssued' topic]
    D --> I['events.TicketPrinted' topic]
    D --> J['events.TicketRefunded' topic]
    D --> K['events.ReadModelIn' topic]
    A -- publish directly --> L['internal-events.svc-tickets.InternalOpsReadModelUpdated'<br>topic]


classDef orange fill:#f96,stroke:#333,stroke-width:4px;
class L orange
```

It will not go through the `events` topic, so it won't be stored to the data lake.

## Exercise

Exercise path: ./project

Create a new event called `InternalOpsReadModelUpdated`. Emit this event after each read model update. It should be published to the `internal-events.svc-tickets.InternalOpsReadModelUpdated` topic.

{{hints}}

{{hint 1}}

To change the topics used for publishing and subscribing, 
you need to adjust the {{exerciseLink "event bus" "09-cqrs-events" "01-event-bus"}} config and {{exerciseLink "event processor" "09-cqrs-events" "04-event-processor"}} config.
Both of them have access to the published and subscribed events.

```go
cqrs.EventBusConfig{
     GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
         event, ok := params.Event.(entities.Event)
         if !ok {
             return "", fmt.Errorf("invalid event type: %T doesn't implement entities.Event", params.Event)
         }

         if event.IsInternal() {
             // Publish directly to the per-event topic
             return "internal-events.svc-tickets." + params.EventName, nil
         } else {
             // Publish to the "events" topic, so it will be stored to the data lake and forwarded to the
             // per-event topic
             return "events", nil
         }
     }
```

```go
return cqrs.EventProcessorConfig{
     GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
         handlerEvent := params.EventHandler.NewEvent()
         event, ok := handlerEvent.(entities.Event)
         if !ok {
             return "", fmt.Errorf("invalid event type: %T doesn't implement entities.Event", handlerEvent)
         }

         var prefix string
         if event.IsInternal() {
             prefix = "internal-events.svc-tickets."
         } else {
             prefix = "events."
         }

         return prefix + params.EventName, nil
     },
```

{{endhint}}

{{hint 2}}

How do you publish an event when the read model is updated? Just inject the event bus to the read model and publish the event.

```go
    // ...
	
			return r.updateReadModel(ctx, tx, updatedRm)
		},
	); err != nil {
		return err
	}

	return r.eventBus.Publish(ctx, &entities.InternalOpsReadModelUpdated{
		Header:    entities.NewMessageHeader(),
		BookingID: uuid.MustParse(bookingID),
	})
}
```

{{endhint}}

{{endhints}}
