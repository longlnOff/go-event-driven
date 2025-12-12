# Project: Context logger

You now have the correlation ID propagated everywhere.
It would be helpful to include it in all log messages.

To make it easier to use in any place in code, you can create a logger with a correlation ID field
and keep it in the request's context. The training's [common library](https://github.com/ThreeDotsLabs/go-event-driven) provides the `log.ToContext` function that does this.

Use `slog.With` to create a logger with a correlation ID field that will be added to each logged line.

```go
ctx := log.ToContext(ctx, slog.With("correlation_id", correlationID))
```

There's also `log.FromContext`, which retrieves the logger:

```go
logger := log.FromContext(msg.Context())
logger.With("key", "value").Info("Log message")
```

## Exercise

Exercise path: ./project

1. **Modify the correlation ID middleware you created before to store the logger in the context.**

{{tip}}

Instead of changing the middleware you can also add a separate one that gets the correlation ID out of the context (with `log.CorrelationIDFromContext`)
and stores it in the logger.

Pick the approach you like best.

{{endtip}}

2. **Modify the logger middleware to use the logger from the context to log messages.**

Now, every `Handling a message` log should automatically include a `correlation_id` field.

Remember that **the order of middleware functions is important:** the logger middleware should come after the correlation ID middleware.
