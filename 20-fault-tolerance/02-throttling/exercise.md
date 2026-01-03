# Throttling

Publishing messages is usually a very lightweight operation.
In busy systems, thousands of messages can be published per second.

Handling these messages, on the other hand, is often more expensive.
There's usually some external resource involved, like a database or an API that has its own limits.
Even if you can spawn more subscribers, you may not be able to scale the external resources as easily.

Throttling is a mechanism to limit the number of concurrent requests executing a handler.
It is useful when a database can't handle big spikes of messages that happen a few times a day 
or if you use a third-party API that has a rate limit.

You can throttle your subscribers so that messages are not processed faster than a given threshold.
This can slow down how quickly you process messages, but it will prevent your external resources (like the database) from being overloaded,
which can cause bigger issues across your system.

Watermill provides throttling middleware that's trivial to use.

```go
// Process no more than 1000 messages per minute
t := middleware.NewThrottle(1000, time.Minute)

router.AddMiddleware(t.Middleware)
```

You can also try implementing your own throttling middleware if you're feeling adventurous.

## Exercise

Exercise path: ./20-fault-tolerance/02-throttling/main.go

This exercise shows a system that, in reaction to the `UserSignedUp` event, sends a welcome message to the user via SMS.
We use an external API for the SMS communication with a rate limit of 10 messages per second.
If we go over the limit, the API blocks until we manually contact support.

Add throttling middleware to the router so that we never send more than 10 messages per second.

{{tip}}

You don't need to add it to your project - we are just checking it in this exercise.

If you want to add it to your project, watch out to not make it too aggressive,
so your solution will not time out.

{{endtip}}
