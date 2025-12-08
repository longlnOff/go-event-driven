# Events

For all our event-driven talk so far, we haven't yet look into events; so far, we've been using plain messages.

An event is a kind of message, and in terms of the data going through the Pub/Sub, there's no difference between them.

The most important thing about an event is that it's **about something that already happened**.
It's an immutable fact.
Whatever happens after an event can't change it.

Technically, it's not a huge difference, but conceptually, it reverses the way you design your system.

Consider what happens when you place an order on any e-commerce web app. It could be something like this:

```go
func PlaceOrder(order Order) {
	SaveOrder(order)
	NotifyUser(order)
	NotifySales(order)
	NotifyWarehouse(order)
	GenerateInvoice(order)
	ChargeCustomer(order)
}
```

**The `PlaceOrder` function becomes tightly coupled to all other services that have something to do with the order.**
Even if we use a Pub/Sub, every service is triggered from this single place.

Introducing an event reverses the responsibility.

```go
func PlaceOrder(order Order) {
	SaveOrder(order)
	PublishOrderPlaced(order)
}
```

Now, `PlaceOrder` is responsible only for storing the order and publishing an event about that fact (`OrderPlaced`).
Any other service can subscribe to that event and react however it wants.
We can have even hundreds of subscribers that are independent of the `PlaceOrder` function.

Technically, the two approaches are very similar; the difference is where the logic for reacting to a change is.
All subscribers are decoupled from the publisher, and they can be modified or extended independently of the publisher.
This helps scale the system in terms of both performance and team independence.

**If we want to add another action after placing an order, we can simply subscribe to the existing event,
without needing to change the `PlaceOrder` function.** In fact, it won't even be aware of the new action.

As a bonus, all the events serve as a record of what happened in the system.
When you follow this history, it's easy to understand what's going on.

To sum up, **an event should be a verb in past tense stating that something happened**.
When designing the event, think about what happened, not what needs to happen after.
Otherwise, you may fall into the "passive-aggressive events" trap,
where the publisher knows what happens after the event is published.

| Good event name  | Bad event name                          |
|------------------|-----------------------------------------|
| `OrderPlaced`    | `NotificationOfUserOnOrderShouldBeSent` |
| `UserSignedUp`   | `SendWelcomeEmailIsReadyToSend`         |
| `AlarmTriggered` | `AlarmNeedsToBeEnabled`                 |
