# Message Processing Metrics

We now have a pretty solid base on Prometheus metrics. Let's use it to measure something more meaningful than a dummy metric.

We will add three metrics:

- `messages_processed_total`: The total number of processed messages (counter)
- `messages_processing_failed_total`: The total number of message processing failures (counter)
- `messages_processing_duration_seconds`: The total time spent processing messages (summary with quantiles 0.5, 0.9, and 0.99)

We want to know which topics and handlers are processing the most messages (or failing the most).
We need to add labels `topic` and `handler` to our metrics to know that.

{{tip}}

Please remember to avoid high-cardinality labels because they can lead to high memory usage and performance issues.

In many cases, having a label for each error message is not a good idea. 
If you have an error message with a different message for each request (or message) and they occur often,
this may lead to high memory usage and performance issues — in a worst-case scenario, it can crash your application due to lack of memory.

{{endtip}}

## Exercise

Exercise path: ./project

To add metrics with labels, we need to use the [`NewCounterVec`](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promauto#NewCounterVec) and [`NewSummaryVec`](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promauto#NewSummaryVec) functions.
They are very similar to the [https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promauto#NewCounter](`NewCounter`) and [`NewSummary`](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promauto#NewSummary) functions, but they can handle the same metrics with different labels.

In the metric options, we need to provide the label names we will use.
For the _summary_ metric, we also need to provide how precise we want our quantiles to be.
In the configuration, it's called `Objectives` and is a map of quantiles to their absolute error.

```go
import (
    "tickets/adapters"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ...

messagesProcessingDuration = promauto.NewSummaryVec(
    prometheus.SummaryOpts{
        Namespace:  "messages",
        Name:       "processing_duration_seconds",
        Help:       "The total time spent processing messages",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    },
    []string{"topic", "handler"},
)
```

We can implement all of these metrics in {{exerciseLink "message middleware" "06-middlewares" "01-add-middleware"}}.

You can extract the topic and event name from the message context:

```go
import (
    "tickets/adapters"
    "github.com/ThreeDotsLabs/watermill/message"
)

// ...

topic := message.SubscribeTopicFromCtx(msg.Context())
handler := message.HandlerNameFromCtx(msg.Context())
```

To store the metric with a label, we need to call:

```go
labels := prometheus.Labels{"topic": topic, "handler": handler}

// ...

messagesProcessedCounter.With(labels).Inc()
```

The resulting metrics should look like this:

```text
# HELP messages_processed_total The total number of processed messages
# TYPE messages_processed_total counter
messages_processed_total{handler="AppendToTracker",topic="events.TicketBookingConfirmed_v1"} 1
messages_processed_total{handler="BookPlaceInDeadNation",topic="events.BookingMade_v1"} 1
messages_processed_total{handler="IssueReceipt",topic="events.TicketBookingConfirmed_v1"} 1
messages_processed_total{handler="PrintTicketHandler",topic="events.TicketBookingConfirmed_v1"} 1
messages_processed_total{handler="StoreTickets",topic="events.TicketBookingConfirmed_v1"} 1
messages_processed_total{handler="TicketRefund",topic="commands.RefundTicket"} 1
messages_processed_total{handler="events_forwarder",topic="events_to_forward"} 1
messages_processed_total{handler="events_splitter",topic="events"} 5
messages_processed_total{handler="ops_read_model.IssueReceiptHandler",topic="events.TicketReceiptIssued_v1"} 1
messages_processed_total{handler="ops_read_model.OnBookingMade",topic="events.BookingMade_v1"} 1
messages_processed_total{handler="ops_read_model.OnTicketBookingConfirmed",topic="events.TicketBookingConfirmed_v1"} 1
messages_processed_total{handler="ops_read_model.OnTicketPrinted",topic="events.TicketPrinted_v1"} 1
messages_processed_total{handler="ops_read_model.OnTicketRefunded",topic="events.TicketRefunded_v1"} 1
messages_processed_total{handler="store_to_data_lake",topic="events"} 3

# HELP messages_processing_duration_seconds The total time spent processing messages
# TYPE messages_processing_duration_seconds summary
messages_processing_duration_seconds{handler="AppendToTracker",topic="events.TicketBookingConfirmed_v1",quantile="0.5"} 0.137299958
messages_processing_duration_seconds{handler="AppendToTracker",topic="events.TicketBookingConfirmed_v1",quantile="0.9"} 0.137299958
messages_processing_duration_seconds{handler="AppendToTracker",topic="events.TicketBookingConfirmed_v1",quantile="0.99"} 0.137299958
messages_processing_duration_seconds_sum{handler="AppendToTracker",topic="events.TicketBookingConfirmed_v1"} 0.137299958
messages_processing_duration_seconds_count{handler="AppendToTracker",topic="events.TicketBookingConfirmed_v1"} 1
messages_processing_duration_seconds{handler="BookPlaceInDeadNation",topic="events.BookingMade_v1",quantile="0.5"} 0.218044
messages_processing_duration_seconds{handler="BookPlaceInDeadNation",topic="events.BookingMade_v1",quantile="0.9"} 0.218044
messages_processing_duration_seconds{handler="BookPlaceInDeadNation",topic="events.BookingMade_v1",quantile="0.99"} 0.218044
messages_processing_duration_seconds_sum{handler="BookPlaceInDeadNation",topic="events.BookingMade_v1"} 0.218044
messages_processing_duration_seconds_count{handler="BookPlaceInDeadNation",topic="events.BookingMade_v1"} 1
// ...
```

{{hints}}

{{hint 1}}

To record the duration of message processing, you can call [`time.Now()`](https://pkg.go.dev/time#Now) at the beginning of the middleware and [`time.Since()`](https://pkg.go.dev/time#Since) at the end.

```go
start := time.Now()

// ...

messagesProcessingDuration.With(labels).Observe(time.Since(start).Seconds())
```

{{endhint}}

{{hint 2}}

To find out that message processing failed, it's enough to check if the error from the next handler is not nil.


```go
router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
    return func(msg *message.Message) (events []*message.Message, err error) {
        // ...
		
        msgs, err := h(msg)
        if err != nil {
            messagesProcessingFailedCounter.With(labels).Inc()
        }

        // ...

        return msgs, err
    }
})
```

{{endhint}}

{{endhints}}
