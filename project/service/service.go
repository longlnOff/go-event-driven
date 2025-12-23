package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	ticketsDB "tickets/db"
	ticketsHttp "tickets/http"
	ticketsMessage "tickets/message"
	ticketsEvent "tickets/message/event"
	ticketsOutbox "tickets/message/outbox"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	db            *sqlx.DB
	echoRouter    *echo.Echo
	messageRouter *message.Router
}

func New(
	dbConn *sqlx.DB,
	spreadsheetsAPI ticketsEvent.SpreadsheetsAPI,
	receiptsService ticketsEvent.ReceiptsService,
	fileService ticketsEvent.FilesService,
	rdb redis.UniversalClient,
) Service {

	watermillLogger := watermill.NewSlogLogger(log.FromContext(context.Background()))
	publisher := ticketsMessage.NewRedisPublisher(rdb, watermillLogger)
	eventBus := ticketsEvent.NewEventBus(publisher, watermillLogger)

	ticketRepo := ticketsDB.NewTicketsRepository(dbConn)
	showRepo := ticketsDB.NewShowsRepository(dbConn)
	bookingRepo := ticketsDB.NewBookingRepository(dbConn)

	postgresSubscriber := ticketsOutbox.NewPostgresSubscriber(dbConn, watermillLogger)

	eventHandler := ticketsEvent.NewEventHandler(
		spreadsheetsAPI,
		receiptsService,
		fileService,
		ticketRepo,
		eventBus,
	)
	eventProcessorConfig := ticketsEvent.NewEventProcessorConfig(
		rdb,
		watermillLogger,
	)
	router := ticketsMessage.NewRouter(
		postgresSubscriber,
		publisher,
		*eventProcessorConfig,
		eventHandler,
		watermillLogger,
	)

	echoRouter := ticketsHttp.NewHttpRouter(
		eventBus,
		ticketRepo,
		showRepo,
		bookingRepo,
	)

	return Service{
		db:            dbConn,
		echoRouter:    echoRouter,
		messageRouter: router,
	}
}

func (s Service) Run(ctx context.Context) error {
	if err := ticketsDB.InitializeDatabaseSchema(s.db); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}
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
