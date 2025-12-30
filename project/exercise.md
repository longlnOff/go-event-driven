# Store events in Data Lake

Now, when we have all of our events on a single topic, we can store them in the data lake.
For the purposes of the training, we will use PostgreSQL as our data lake.

{{tip}}

PostgreSQL is not ideal for very large scale data lakes (like terabytes of data).
Large datasets make PostgreSQL too expensive for storing all events.

{{endtip}}

You can think about a data lake as being like a big {{exerciseLink "read model" "13-read-models" "01-read-models"}} containing all events in raw form.

## Exercise

Exercise path: ./project

1. Create a new database table called `events`.

**It's important to use exactly this schema:**

```sql
CREATE TABLE IF NOT EXISTS events (
    event_id UUID PRIMARY KEY,
    published_at TIMESTAMP NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    event_payload JSONB NOT NULL
);
```

2. Add {{exerciseLink "a new message handler" "04-router" "01-handlers"}} that stores all events in the data lake.

It should listen to the `events` topic.

{{tip}}

Don't forget to use a separate {{exerciseLink "consumer group" "03-message-broker" "05-consumer-groups"}} for this handler!

{{endtip}}

3. Store all events in the `events` table.

Get `event_id` and `published_at` from the event header. You should unmarshal it from the event's payload.
Extract `event_name` from the message using the CQRS marshaler (`cqrs.JSONMarshaler`).

`event_payload` should contain the raw event payload of the message (`msg.Payload`).

An event handler won't work here because we need to store our events in raw form.

Don't forget about {{exerciseLink "at-least-once delivery" "10-at-least-once-delivery" "06-project-handling-re-delivery"}}!
You should deduplicate potential redelivered events based on the `event_id`.

{{hints}}

{{hint 1}}
To store events in a data lake, extract `event_id` and `published_at` from the event header.
Unmarshal the event to a struct with just a header field. The rest of the payload will be ignored:

```go
// We just need to unmarshal the event header; the rest is stored as is.
type Event struct {
	Header entities.MessageHeader `json:"header"`
}

var event Event
if err := eventProcessorConfig.Marshaler.Unmarshal(msg, &event); err != nil {
	return fmt.Errorf("cannot unmarshal event: %w", err)
}
```

{{endhint}}

{{hint 2}}

As in previous exercises, you can extract the event name from a message by using the CQRS marshaler:

```go
eventName := eventProcessorConfig.Marshaler.NameFromMessage(msg)
if eventName == "" {
	return fmt.Errorf("cannot get event name from message")
}
```

{{endhint}}

{{endhints}}
