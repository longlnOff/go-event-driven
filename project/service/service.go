package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	ticketsHttp "tickets/http"
	ticketsMessage "tickets/message"
)

type Service struct {
	echoRouter *echo.Echo
}

func New(
	spreadsheetsAPI ticketsMessage.SpreadsheetsAPI,
	receiptsService ticketsMessage.ReceiptsService,
	rdb redis.UniversalClient,
) Service {

	watermillLogger := watermill.NewSlogLogger(nil)
	publisher := ticketsMessage.NewRedisPublisher(rdb, watermillLogger)

	ticketsMessage.NewHandler(
		spreadsheetsAPI,
		receiptsService,
		rdb,
		watermillLogger,
	)

	echoRouter := ticketsHttp.NewHttpRouter(
		publisher,
	)
	
	return Service{
		echoRouter: echoRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	err := s.echoRouter.Start(":8080")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
