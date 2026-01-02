# Tracing

{{background}}

We now know how to approach building a production-grade alerting for our system.
We can't wait to be woken up in the middle of the night by an alert!

When it finally happened to one of our colleagues, he was eager to fix the issue.
It turned out that the `BookingMade_v1` event was stuck, and bookings were not propagating to Dead Nation.

Before fixing the issue, he wanted to understand the context of the problem a bit more.
Unfortunately, he joined our team two weeks ago and didn't know all the decisions that we made and how everything was connected.
He was looking for that until 8 AM, when the first team members started to wake up.
We probably shouldn't have included him in the off-hours on-call rotation.
And we should also have some tools to help him understand how the system works.

{{endbackground}}

It's hard to understand how distributed systems work, especially integrated via events.
One of the best things that we can do to improve the observability of our system is to add **distributed tracing**.

Tracing automatically generates a trace of all requests and messages that go through the system, even if they flow through multiple services.
What's more important, they will be correlated together.
In other words, by knowing the trace ID connected to an event, we can see what operation published the event and what the event triggered.

For our ticketing system, we want to achieve something like this:

{{img "Tickets trace" "tickets-trace.png"}}

In this module, we'll use the [OpenTelemetry](https://opentelemetry.io) library.
If you've had a chance to work with OpenCensus or OpenTracing, it's their successor that is the result of merging those two projects.

OpenTelemetry provides multiple exporters, and they can be easily replaced.
For storing traces, we'll use the open-source [Jaeger](https://www.jaegertracing.io).
In your work projects, you can use Jaeger locally (with Docker) and a cloud-based solution (like AWS X-Ray or GCP Stack Driver) on production.

## Tracing Glossary

### Trace

A group of multiple spans with the same trace ID.

### Span

A single operation within a trace. It's defined by the following attributes:

- Span ID
- Operation name
- Start and end times
- Key/value attributes
- Parent's span ID

{{img "Span vs Trace" "span-vs-trace.png"}}

Example span attributes:

{{img "Span attributes" "span-attributes.png"}}

Example operations that can be represented as spans:

- HTTP request
- Database query
- Processing a message
- Sending a message

In general, spans work best when they have a meaningfully long duration.
They are not meant to be used for very short operations, like a single function call.

### Exporter

An implementation for exporting traces that is responsible for sending traces to Jaeger, AWS X-Ray, GCP Stack Driver, etc.

### Propagator

Used to serialize and deserialize trace context. It's used to pass trace context between services.
We will cover how to use it in detail in later exercises.

### Sampling

Sampling is used to reduce the number of traces that are collected and sent to the backend.
Usually, we don't want to trace all requests because that will generate too much data.

You can find a more detailed description of the glossary used in OpenTelemetry (and in tracing in general) here:

- https://opentelemetry.io/docs/concepts/
- https://opentelemetry.io/docs/concepts/glossary/

## Exercise

Exercise path: ./19-tracing/01-tracing/main.go

Let's add a simple tracing to the `FindUser` and `AddUser` functions.

To start a new span, you need to call:

```go
import "go.opentelemetry.io/otel"

ctx, span := otel.Tracer("").Start(ctx, "AddUser")
defer span.End()
```

`Start` automatically creates a new span and adds it to the context.
If the context passed to `Start` already contains a span, it will be used as a parent span.
In the end, they will be connected within one trace.

Calling `End()` on the span marks it as finished.
It often works best to call it with `defer`, so you can be sure that it will be called even if the function returns early.

{{tip}}

You may ask: What about the `""` in `otel.Tracer("")`?

This is supposed to be a name of the tracer. As long as you are not implementing a library, it can be empty.
If you are implementing a library, you should use the name of your library.
If you are creating your common library for services, you can also put your service name there.
For simple applications, it's totally fine to just keep it empty.

We see it as a bit of unfortunate API design, but it's not a big deal. You can always add a helper to hide it in your project.

{{endtip}}

In our example, you can see that `FindUser` is called by `AddUser`.

{{img "Exercise trace" "simple-trace.png"}}

In this exercise, you don't need to worry about configuring exporters: Tests will handle that.

**Your task is to record spans for the `AddUser` and `FindUser` functions.**
Span names should be equal to `AddUser` and `FindUser`, respectively.

{{tip}}

There is no convention for how to name spans.
It's good to be consistent in your project and use names that will be easy to understand for other developers.

There are no character restrictions in span names, so you can use spaces, dots, etc.

Unlike with metrics, you don't need to watch out for names with high cardinality.

{{endtip}}
