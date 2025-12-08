# Marshalling

So far, our messages have contained just a single string, like the ID.
In real applications, you often need to send more complex data.

Most events can be represented as structs, for example:

```go
type OrderPlaced struct {
	ID         string
	CustomerID string
	Total      Money
	PlacedAt   time.Time
}
```

To send this over the Pub/Sub, you need to marshal (serialize) it to a slice of bytes.
For example, using JSON:

```go
type OrderPlaced struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Total      string `json:"total"`
	PlacedAt   string `json:"placed_at"`
}
```

Notice how we've changed complex types (`Money` and `time.Time`) to primitive types.

You can follow two paths here: Either keep complex types and use their default marshallers
(or define custom marshallers for them), or keep only primitive types in the events.

The first approach is more convenient, but gives you less control over the output.
For example, `time.Time` marshals to a string in the `RFC3339Nano` format by default.
If you need to change it at some point, you will need to rework all the fields that use it. 

Using only primitive types makes the result more explicit, but requires more boilerplate.
You would usually keep two types for each event: one on the application layer, and one on the Pub/Sub layer.
You would need to map the fields manually between them.

Consider which approach is better for your use case.
As a rule of thumb, keeping separate structs can be a good idea if you have a proper domain layer with isolated domain logic.
On the other hand, if your events have lots of fields and aren't close to the domain, the default marshallers may be good enough.

Finally, if you're not sure, start with one approach and make sure it's easy to change later.
Sometimes it's better to have a working solution than to spend too much time on finding the perfect one.

{{tip}}

For more details on using separate models, see our blog posts:

- [When to avoid DRY in Go](https://threedots.tech/post/things-to-know-about-dry/)
- [Introducing Clean Architecture by refactoring a Go project](https://threedots.tech/post/introducing-clean-architecture/)

{{endtip}}


Publishing a marshalled event can look like this:

```go
func PublishOrderPlaced(orderPlaced app.OrderPlaced) error {
	event := OrderPlaced {
		ID:         orderPlaced.ID, 
		CustomerID: orderPlaced.CustomerID, 
		Total:      fmt.Sprintf("%v %v", orderPlaced.Total.Amount, orderPlaced.Total.Currency), 
		PlacedAt:   orderPlaced.PlacedAt.Format(time.RFC3339),
	}
	
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	
	msg := message.NewMessage(watermill.NewUUID(), payload)
	
	return publisher.Publish("orders-placed", msg)
}
```

The exact format doesn't matter that much as long as the publisher and the subscriber use the same one.
While JSON is very popular, there are more specialized formats like [Protocol Buffers](https://protobuf.dev) or [Avro](https://avro.apache.org).

We often use Protocol Buffers for the benefits it provides over JSON:

- It has a typed schema.
- It's easy to generate ready-to-use structs.
- It can be marshaled to a binary format (resulting in smaller messages and faster processing, if that matters for you).
- It can also be marshaled to JSON, making it easier to debug.

As the trade-off you need to set up the tools to generate the code.

Some brokers provide a *schema registry* that allows you to store the message schemas in a central place.
They are beyond the scope of this training.
For Protocol Buffers, a good start is keeping the schema files in a repository.

## Exercise

Exercise path: ./05-events/02-marshalling/main.go

**Add a new handler that subscribes to events from the `payment-completed` topic.**

The handler should unmarshal the message payload to the `PaymentCompleted` struct and
then send a new payload to the `order-confirmed` topic in the JSON format:

```json
{
  "order_id": "...",
  "confirmed_at": "..."
}
```

The `confirmed_at` field should be taken from the `CompletedAt` field of the `PaymentCompleted` event.
