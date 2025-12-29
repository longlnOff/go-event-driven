# Ordering: Independent Updates

The various approaches to ordering can seem complex at first.
Thankfully, you often can choose a simpler way.

A classic example is when you have a single entity that is updated by multiple event handlers,
such as an `Order` read model that has a `status` field.
The status can be updated by multiple events, like `OrderPaid`, `OrderShipped`, `OrderRefunded`, etc.
If one of the events arrives out of order, the status will be overwritten with a previous value.

The solution to this problem is straightforward: Make each event update some fields of the read model independently.
In the example above, you can have an `OrderPaid` event that only updates the `paid_at` field,
an `OrderShipped` event that updates the `shipped_at` field, and an `OrderRefunded` event that updates the `refunded_at` field.

If they arrive out of order, the read model keeps the correct state.
If you need a single field that contains the current status, you can calculate it on the fly when reading the data.
For example, you can compare the times and pick the status based on the latest one.

You can use a similar approach in other scenarios.
The point is to not overwrite any data that could be updated by other events.

## Exercise

Exercise path: ./14-message-ordering/05-independent-updates/main.go

The exercise shows an incident-detection system that receives two events: `AlertTriggered` and `AlertResolved`.

In reaction to them, it publishes an `AlertUpdated` event (a read model for the UI) with the `IsTriggered` field filled in.
If the events come out of order, the `IsTriggered` field can contain an incorrect value.

**Change the `AlertUpdated` event payload to include both times as `last_triggered_at` and `last_resolved_at` fields.**

This way, it's always possible to figure out if the alert is currently triggered by comparing the two times.
