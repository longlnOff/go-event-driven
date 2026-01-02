# Tracing Beyond Messages

We focus on tracing in the context of messages. But there are other topics worth considering in your projects.

## Exposing the trace ID

It's useful to be able to know the trace ID of the request.

It's good enough to just log it somewhere, so you're able to correlate logs with traces.
You can also add it to each log line as a field (but this is not recommended locally — it will pollute your logs a lot).
You can extract the trace ID from the context by calling [`trace.SpanContextFromContext(ctx).TraceID().String()`](https://pkg.go.dev/go.opentelemetry.io/otel/trace#SpanContextFromContext).

You can also use a trick here: Use the trace ID as a correlation ID.
This will let you simplify your service logic and remove the propagation of the correlation ID.
Of course, to do that, you need to adjust your logging logic.

## Adding HTTP requests to traces

You can achieve this with the [`go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`](go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp) library.

You need to use a custom HTTP client for your API clients:

```diff
 import (
    "tickets/adapters"
 	"context"
+	"fmt"
 	"net/http"
 	"os"
 	"os/signal"
 	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
 	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
 	"github.com/jmoiron/sqlx"
+	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
+	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
 )
 
 func main() {
 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
 	defer cancel()
 
-	apiClients, err := clients.NewClients(
+	traceHttpClient := &http.Client{Transport: otelhttp.NewTransport(
+		http.DefaultTransport,
+		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
+			return fmt.Sprintf("HTTP %s %s %s", r.Method, r.URL.String(), operation)
+		}),
+	)}
+
+	apiClients, err := clients.NewClientsWithHttpClient(
 		os.Getenv("GATEWAY_ADDR"),
 		func(ctx context.Context, req *http.Request) error {
 			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
 			return nil
 		},
+		traceHttpClient,
 	)
```

This will automagically add all outgoing HTTP requests to traces.
It would also add the `traceparent` header to all outgoing HTTP requests, but we call external APIs only, so it won't be used.

However, if you call your own services, you'll be able to see them all in traces.

## Adding SQL queries to traces

It's nice to have SQL queries in traces, especially when debugging some performance issues.

To do that, you need to wrap your SQL connection with [`github.com/uptrace/opentelemetry-go-extra/otelsql`](github.com/uptrace/opentelemetry-go-extra/otelsql).

```diff
 import (
    "tickets/adapters"
 	"context"
 	"net/http"
 	"os"
 	"os/signal"
@@ -14,24 +15,43 @@ import (
    "tickets/adapters"
 	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
 	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
 	"github.com/jmoiron/sqlx"
+	"github.com/uptrace/opentelemetry-go-extra/otelsql"
 )
 

-	db, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
+	traceDB, err := otelsql.Open("postgres", os.Getenv("POSTGRES_URL"),
+		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
+		otelsql.WithDBName("db"))
+	if err != nil {
+		panic(err)
+	}
+
+	db := sqlx.NewDb(traceDB, "postgres")
 	if err != nil {
 		panic(err)
 	}
```

As long as you pass the context to all your SQL queries (`ExecContext`, `NamedExecContext`, etc.), they will be added to traces.

## Exercise

Exercise path: ./project

We won't check your solution for this, but we recommend adding tracing for HTTP requests and SQL queries.

We'll run the same tests as in the previous exercise, but you can also run your service locally and check the traces in Jaeger.
