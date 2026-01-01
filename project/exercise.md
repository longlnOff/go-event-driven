# Migrating the Read Model

Remember we wanted to learn how to {{exerciseLink "migrate our read models" "15-data-lake" "01-data-lake"}}?
We now have all the building blocks.
It's time to put them together!

{{tip}}

The {{exerciseLink "data lake module" "15-data-lake" "01-data-lake"}} was optional.
You may have skipped it.

If you want to revisit, use `tdl tr jump` to jump to the module and complete it now.

{{endtip}}

Usually, to migrate the read model, you need to follow these steps:

1. Query events from the data lake one by one, from oldest to newest.
2. If needed, do a mapping of versions (your read model may be built from newer versions of events, 
while in the data lake, you may have older versions).
3. Call your read model methods to build it. 
   Usually, it will be some form of a repository similar to what you implemented in {{exerciseLink "13-read-models/01-read-models" "13-read-models" "01-read-models"}}.
   You can simplify your life by calling read model repository methods directly, not via message handlers.
   You don't need to publish a message or event. Just call the repository method directly.

It's how we'll migrate our read model in this exercise.

{{tip}}

Usually, if the migration is long-running, you may want to do the migration in the background and have some resume mechanism. 
For example, you can store the last timestamp of the event that you processed and
start from that timestamp when you resume the migration.

In our case, it will be not needed.

{{endtip}}

{{tip}}

What if you needed to build a new read model but didn't have a data lake?
You can always do some reverse-engineering and build a read model based on your {{exerciseLink "write model" "13-read-models" "01-read-models"}}.
You can query your production tables (like the `bookings` table) and build a read model from that data.

{{endtip}}

## Exercise

Exercise path: ./project

Ensure the `events` table has the same schema as in the {{exerciseLink "15-data-lake/03-project-store-events-to-data-lake" "15-data-lake" "03-project-store-events-to-data-lake"}} exercise.

1. **Wait for us to insert the old events to the `events` table.** 
We will insert events there after the table is created.
**We insert all events at once. You can safely assume that if the `events` table is not empty, you can migrate your read model.**

**Do not start migration before we populate the `events` table.**
You can run the migration in a goroutine. 
Put it in a function that runs your service or in the `main` function.

2. Iterate over all events in the `events` table. **Start from the oldest one.**

3. Unmarshal events from the data lake to your events format.

**We will publish events with version `v0` since we don't know your exact event format.**

The events that we will add to the `events` table are:
- `BookingMade_v0`
- `TicketBookingConfirmed_v0`
- `TicketReceiptIssued_v0`
- `TicketRefunded_v0`
- `TicketPrinted_v0`

In a format:

```go
type bookingMade_v0 struct {
	Header entities.MessageHeader `json:"header"`

	NumberOfTickets int `json:"number_of_tickets"`

	BookingID uuid.UUID `json:"booking_id"`

	CustomerEmail string    `json:"customer_email"`
	ShowID        uuid.UUID `json:"show_id"`
}

type ticketBookingConfirmed_v0 struct {
	Header entities.MessageHeader `json:"header"`

	TicketID      string         `json:"ticket_id"`
	CustomerEmail string         `json:"customer_email"`
	Price         entities.Money `json:"price"`

	BookingID string `json:"booking_id"`
}

type ticketReceiptIssued_v0 struct {
	Header entities.MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	ReceiptNumber string `json:"receipt_number"`

	IssuedAt time.Time `json:"issued_at"`
}

type ticketPrinted_v0 struct {
	Header entities.MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

type ticketRefunded_v0 struct {
	Header entities.MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}
```

4. Map `v0` events to your `v1` events. Migrate your Ops read model (`read_model_ops_bookings`) based on events from the data lake.

Call your read model methods directly.
For example:

```go
bookingConfirmedEvent, err := unmarshalDataLakeEvent[ticketBookingConfirmed_v0](event)
if err != nil {
   return err
}

return rm.OnTicketBookingConfirmed(ctx, &entities.TicketBookingConfirmed_v1{
   // you should map v0 event to your v1 event here
   Header:        bookingConfirmedEvent.Header,
   TicketID:      bookingConfirmedEvent.TicketID,
   CustomerEmail: bookingConfirmedEvent.CustomerEmail,
   Price:         bookingConfirmedEvent.Price,
   BookingID:     bookingConfirmedEvent.BookingID,
})
```

**We'll verify your solution by querying the HTTP endpoint to get data from the Read Model.**

{{tip}}

In the real world, you may not need to build a read model from the oldest data.
You may decide on some cut-off date and build the read model only from that date.
It depends on the use case, but if you are going to build it from the newest data, mapping may be not needed.

{{endtip}}

{{tip}}

We recommend adding a good amount of logs to your migration.
It's not unusual that such migrations take longer than expected or something goes wrong.
It's worth logging the progress, how many events were processed, how much time it took, etc.

{{endtip}}

{{hints}}

{{hint 1}}

This is an example method in the data lake repository. It gets all events in the correct order.

```go

type DataLakeEvent struct {
	EventID      string    `db:"event_id"`
	PublishedAt  time.Time `db:"published_at"`
	EventName    string    `db:"event_name"`
	EventPayload []byte    `db:"event_payload"`
}

// ...

func (d DataLake) GetEvents(ctx context.Context) ([]entities.DataLakeEvent, error) {
	var events []entities.DataLakeEvent
	err := d.db.SelectContext(ctx, &events, "SELECT * FROM events ORDER BY published_at ASC")
	if err != nil {
		return nil, fmt.Errorf("could not get events from data lake: %w", err)
	}

	return events, nil
}
```

{{endhint}}

{{hint 2}}

Here's an example generic function to unmarshal a data lake event into a struct:

```go
func unmarshalDataLakeEvent[T any](event entities.DataLakeEvent) (*T, error) {
	eventInstance := new(T)

	err := json.Unmarshal(event.EventPayload, &eventInstance)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event %s: %w", event.EventName, err)
	}

	return eventInstance, nil
}
```

{{endhint}}

{{hint 3}}

Here's an example of mapping events from `v0` to `v1` and calling the read model:

```go
func migrateEvent(ctx context.Context, event entities.DataLakeEvent, rm db.OpsBookingReadModel) error {
	switch event.EventName {
	case "BookingMade_v0":
		bookingMade, err := unmarshalDataLakeEvent[bookingMade_v0](event)
		if err != nil {
			return err
		}

		return rm.OnBookingMade(ctx, &entities.BookingMade_v1{
			// you should map v0 event to your v1 event here
			Header:          bookingMade.Header,
			NumberOfTickets: bookingMade.NumberOfTickets,
			BookingID:       bookingMade.BookingID,
			CustomerEmail:   bookingMade.CustomerEmail,
			ShowID:          bookingMade.ShowID,
		})
	case "TicketBookingConfirmed_v0":
		bookingConfirmedEvent, err := unmarshalDataLakeEvent[ticketBookingConfirmed_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketBookingConfirmed(ctx, &entities.TicketBookingConfirmed_v1{
			// you should map v0 event to your v1 event here
			Header:        bookingConfirmedEvent.Header,
			TicketID:      bookingConfirmedEvent.TicketID,
			CustomerEmail: bookingConfirmedEvent.CustomerEmail,
			Price:         bookingConfirmedEvent.Price,
			BookingID:     bookingConfirmedEvent.BookingID,
		})
	case "TicketReceiptIssued_v0":
		receiptIssuedEvent, err := unmarshalDataLakeEvent[ticketReceiptIssued_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketReceiptIssued(ctx, &entities.TicketReceiptIssued_v1{
			// you should map v0 event to your v1 event here
			Header:        receiptIssuedEvent.Header,
			TicketID:      receiptIssuedEvent.TicketID,
			ReceiptNumber: receiptIssuedEvent.ReceiptNumber,
			IssuedAt:      receiptIssuedEvent.IssuedAt,
		})
	case "TicketPrinted_v0":
		ticketPrintedEvent, err := unmarshalDataLakeEvent[ticketPrinted_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketPrinted(ctx, &entities.TicketPrinted_v1{
			// you should map v0 event to your v1 event here
			Header:   ticketPrintedEvent.Header,
			TicketID: ticketPrintedEvent.TicketID,
			FileName: ticketPrintedEvent.FileName,
		})
	case "TicketRefunded_v0":
		ticketRefundedEvent, err := unmarshalDataLakeEvent[ticketRefunded_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketRefunded(ctx, &entities.TicketRefunded_v1{
			// you should map v0 event to your v1 event here
			Header:   ticketRefundedEvent.Header,
			TicketID: ticketRefundedEvent.TicketID,
		})
	default:
		return fmt.Errorf("unknown event %s", event.EventName)
	}
}
```

{{endhint}}

{{endhints}}
