package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"

	ticketsAdapter "tickets/adapters"
	ticketsMessage "tickets/message"
	ticketsService "tickets/service"
)

func main() {
	log.Init(slog.LevelInfo)

	apiClients, err := clients.NewClients(os.Getenv("GATEWAY_ADDR"), nil)
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	spreadsheetsAPI := ticketsAdapter.NewSpreadsheetsAPIClient(apiClients)
	receiptsService := ticketsAdapter.NewReceiptsServiceClient(apiClients)

	rdb := ticketsMessage.NewRedisClient(os.Getenv("REDIS_ADDR"))


	err = ticketsService.New(
		spreadsheetsAPI,
		receiptsService,
		rdb,
	).Run(ctx)
	if err != nil {
		panic(err)
	}
}
