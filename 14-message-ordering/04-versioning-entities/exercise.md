# Ordering: Versioning Entities
Another technique to enforce message processing order is entity versioning. It's useful when Pub/Sub has no ordering guarantees.

**Entity versioning generates a version number for an entity during every update and includes it in the event. When processing the event, we check if the version number equals the expected current version of the entity.** If we receive an event with an unexpected version, we can reject it by nacking and wait for another event to arrive.

If we receive continuous data and only care about the latest data, we can ignore older versions when we already have newer data.

Consider this event that we mentioned in one of the previous exercises:

```go
type ProductUpdated struct {
	ID          string
	Name        string
	Description string
	Price       string
}
```

As discussed earlier, if this event is processed out of order, the older changes may overwrite the newer ones.

The solution is to version the entity, in this case, the `Product`.

We can add a `Version` field to the event:

```go
type ProductUpdated struct {
	ID          string
	Name        string
	Description string
	Price       string
	
	Version int
}
```

When handling the event, the handler checks if the version number is equal to the current version of the product in the database plus one.
If it's bigger, it means that the event is out of order, and the handler should return it back to the queue.

While updating the product, the handler should also update the version of the product in the database.

{{tip}}

Message metadata might be also a good place to store the version of the entity.

{{endtip}}

## Exercise

Exercise path: ./14-message-ordering/04-versioning-entities/main.go

The exercise shows an incident-detection system that receives two events: `AlertTriggered` and `AlertResolved`.
In reaction to them, it publishes an `AlertUpdated` event (a read model for the UI) with the `IsTriggered` field filled in.
If the events come out of order, the `IsTriggered` field can contain an incorrect value.

1. Add a `LastAlertVersion` field to the `AlertUpdated` event, marshaling as JSON to `last_alert_version`.

2. Set the `LastAlertUpdated` field in the `AlertUpdated` event based on the incoming event. The versioning starts from `1` and is always present in `AlertTriggered` and `AlertResolved` events.

3. When handling events, check if the incoming event has the expected version. For example, if the last stored version is `4`, the next valid event should have version `5`. If the stored alert has version `1` and we receive an event with version `3`, we should return an error and wait for the event with version `2` to arrive.

Check this by verifying if the event version equals the stored alert version plus one. If not, ack the event by returning `nil` from the handler.

{{hints}}

{{hint 1}}

It's how the example check for the version of the event looks like:

```go
if alert.LastAlertVersion+1 != event.AlertVersion {
    logger.Info(fmt.Sprintf("Invalid version: %v (expected %v)", event.AlertVersion, alert.LastAlertVersion+1), nil)
    return nil
}
```

{{endhint}}

{{endhints}}
