package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	ticketsHttp "tickets/http"
	ticketsMessage "tickets/message"
	ticketsEvent "tickets/message/event"
)

type Service struct {
	echoRouter    *echo.Echo
	messageRouter *message.Router
}

func New(
	spreadsheetsAPI ticketsEvent.SpreadsheetsAPI,
	receiptsService ticketsEvent.ReceiptsService,
	rdb redis.UniversalClient,
) Service {

	watermillLogger := watermill.NewSlogLogger(log.FromContext(context.Background()))
	publisher := ticketsMessage.NewRedisPublisher(rdb, watermillLogger)
	eventHandler := ticketsEvent.NewEventHandler(
		spreadsheetsAPI,
		receiptsService,
	)
	router := ticketsMessage.NewRouter(
		eventHandler,
		rdb,
		watermillLogger,
	)

	echoRouter := ticketsHttp.NewHttpRouter(
		publisher,
	)

	return Service{
		echoRouter:    echoRouter,
		messageRouter: router,
	}
}

func (s Service) Run(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		return s.messageRouter.Run(ctx)
	})
	errGroup.Go(func() error {
		<-s.messageRouter.Running()

		err := s.echoRouter.Start(":8080")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	errGroup.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(context.Background())
	})

	return errGroup.Wait()
}
