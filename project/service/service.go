package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	ticketsHttp "tickets/http"
	ticketsMessage "tickets/message"
)

type Service struct {
	echoRouter *echo.Echo
	messsageRouter *message.Router
}

func New(
	spreadsheetsAPI ticketsMessage.SpreadsheetsAPI,
	receiptsService ticketsMessage.ReceiptsService,
	rdb redis.UniversalClient,
) Service {

	watermillLogger := watermill.NewSlogLogger(nil)
	publisher := ticketsMessage.NewRedisPublisher(rdb, watermillLogger)

	router := ticketsMessage.NewRouter(
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
		messsageRouter: router,
	}
}

func (s Service) Run(ctx context.Context) error {
	errgrp, ctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		return s.messsageRouter.Run(ctx)
	})
	errgrp.Go(func() error {
		<- s.messsageRouter.Running()
		
		err := s.echoRouter.Start(":8080")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
 		}
		return nil
	})
	errgrp.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(context.Background())
	})
 
	return errgrp.Wait()
}
