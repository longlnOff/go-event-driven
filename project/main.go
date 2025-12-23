package main

import (
	"context"
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
)

func main() {
	log.Init(slog.LevelInfo)

	db, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	apiClients, err := clients.NewClients(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
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

	rdb := ticketsMessage.NewRedisClient(os.Getenv("REDIS_ADDR"))

	err = ticketsService.New(
		db,
		spreadsheetsAPI,
		receiptsService,
		fileService,
		deadNationService,
		rdb,
	).Run(ctx)
	if err != nil {
		panic(err)
	}
}
