# Tracing attributes

Spans have an additional superpower: They can have attributes.
These are key/value pairs attached to the span.
They can store any context useful to understand what's happening in the span.

For example, an attribute can be a request URL, user ID, or returned HTTP status.
High-cardinality attributes are not a problem.

{{img "Tracing attributes" "trace-attributes.png"}}

## Adding attributes

You can add attributes when creating the span:

```go
import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// ...

ctx, span := otel.Tracer("").Start(
	ctx,
	"AddUser",
	trace.WithAttributes(
		attribute.String("userID", userID),
	),
)
```

You can also add them later on:


```go
span.SetAttributes(
	attribute.String("userID", userID),
)
```

## Error attributes

Spans also have special attributes for errors.
You can set an error message and status code.
This means the tracing system can show you errors in the UI:

```go
import "go.opentelemetry.io/otel/codes"

// ...

defer func() {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}()
```

## Exercise

Exercise path: ./19-tracing/02-tracing-attributes/main.go

1. Add `userID` attribute to the `AddUser` and `FindUser` spans.
2. Record any errors in both functions' spans. Call `span.RecordError()` and `span.SetStatus(codes.Error, err.Error())`.
