# Using Event Processor in the project

Now it's time to use the Event Processor in your project.

## Exercise

Exercise path: ./project

Update your project to use Event Processor instead of the raw Router.

Here are some tips on how to do this:

1. Replace the Router handlers with an EventProcessor and EventHandlers.
2. **Remember to create a new subscriber instance `SubscriberConstructor`** so each handler will use a separate subscriber. 
   Like you did in the {{exerciseLink "the previous exercise" "09-cqrs-events" "06-cqrs-with-consumer-groups"}}.
   If you don't, your message will be processed by only one handler.

You should not do any JSON unmarshaling yourself. Just pass `JSONMarshaler` to the EventProcessor.

After these changes, you should have much less boilerplate code in the project.
You can also be sure that all marshaling and topic topology is consistent across the project.

{{tip}}

Do you remember how, in {{exerciseLink "the errors module" "07-errors" "03-project-malformed-messages"}}, we ignored malformed messages with the wrong message type?
We no longer use this metadata, so if this check is still in your code, it will cause all messages to be ignored.
Make sure that you don't have code like this in your project:

```go
if msg.Metadata.Get("type") != "booking.created" {
	slog.Error("Invalid message type")
	return nil
}
```

{{endtip}}

{{tip}}

If your service doesn't work as expected, double-check that all components use the correct topics.
Logs should be useful for debugging.
You may want to increase the log level to `debug` or `trace` to see more.   

```go
log.Init(watermill.LevelTrace)
```

{{endtip}}
