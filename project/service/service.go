package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/labstack/echo/v4"

	ticketsWorker "tickets/worker"
	ticketsHttp "tickets/http"
)

type Service struct {
	echoRouter *echo.Echo
}

func New(
	spreadsheetsAPI ticketsWorker.SpreadsheetsAPI,
	receiptsService ticketsWorker.ReceiptsService,
) Service {
	ctx := context.Background()
	worker := ticketsWorker.NewWorker(
		spreadsheetsAPI,
		receiptsService,
	)

	go worker.Run(ctx)

	echoRouter := ticketsHttp.NewHttpRouter(
		worker,
	)
	
	return Service{
		echoRouter: echoRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	err := s.echoRouter.Start(":8080")
	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}

	return nil
}
