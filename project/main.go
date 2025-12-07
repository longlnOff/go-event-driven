package main

import (
	"context"
	"log/slog"
	"os"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"


	ticketsAdapter "tickets/adapters"
	ticketsService "tickets/service"
	ticketsMessage "tickets/message"
)

func main() {
	log.Init(slog.LevelInfo)

	apiClients, err := clients.NewClients(os.Getenv("GATEWAY_ADDR"), nil)
	if err != nil {
		panic(err)
	}

	spreadsheetsAPI := ticketsAdapter.NewSpreadsheetsAPIClient(apiClients)
	receiptsService := ticketsAdapter.NewReceiptsServiceClient(apiClients)

	rdb := ticketsMessage.NewRedisClient(os.Getenv("REDIS_ADDR"))


	err = ticketsService.New(
		spreadsheetsAPI,
		receiptsService,
		rdb,
	).Run(context.Background())
	if err != nil {
		panic(err)
	}
}
