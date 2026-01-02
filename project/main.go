package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	ticketsAdapter "tickets/adapters"
	ticketsMessage "tickets/message"
	ticketsService "tickets/service"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func main() {
	log.Init(slog.LevelInfo)

	traceDB, err := otelsql.Open(
		"postgres", os.Getenv("POSTGRES_URL"),
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBName("db"),
	)
	if err != nil {
		panic(err)
	}

	db := sqlx.NewDb(traceDB, "postgres")
	if err != nil {
		panic(err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}()

	traceHttpClients := &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanNameFormatter(
				func(operation string, r *http.Request) string {
					return fmt.Sprintf("HTTP %s %s %s", r.Method, r.URL.String(), operation)
				},
			),
		)}

	apiClients, err := clients.NewClientsWithHttpClient(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
		traceHttpClients,
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	spreadsheetsAPI := ticketsAdapter.NewSpreadsheetsAPIClient(apiClients)
	receiptsService := ticketsAdapter.NewReceiptsServiceClient(apiClients)
	fileService := ticketsAdapter.NewFileServiceClient(apiClients)
	deadNationService := ticketsAdapter.NewDeadNationServiceClient(apiClients)
	paymentsService := ticketsAdapter.NewPaymentsServiceClient(apiClients)

	rdb := ticketsMessage.NewRedisClient(os.Getenv("REDIS_ADDR"))

	err = ticketsService.New(
		db,
		spreadsheetsAPI,
		receiptsService,
		fileService,
		paymentsService,
		deadNationService,
		rdb,
	).Run(ctx)

	if err != nil {
		panic(err)
	}
}
