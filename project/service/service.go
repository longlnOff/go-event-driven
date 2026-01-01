package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	ticketsDB "tickets/db"
	ticketsHttp "tickets/http"
	ticketsMessage "tickets/message"
	ticketsCommand "tickets/message/command"
	ticketsEvent "tickets/message/event"
	ticketsOutbox "tickets/message/outbox"
	readModelMigration "tickets/migrate_read_model"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
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
	opsReadModel  ticketsDB.OpsBookingReadModel
	eventRepo     ticketsDB.EventsRepository
}

type ReceiptService interface {
	ticketsEvent.ReceiptsService
	ticketsCommand.ReceiptsService
}

func New(
	dbConn *sqlx.DB,
	spreadsheetsAPI ticketsEvent.SpreadsheetsAPI,
	receiptsService ReceiptService,
	fileService ticketsEvent.FilesService,
	paymentService ticketsCommand.PaymentsService,
	deadNationService ticketsEvent.DeadNationService,
	rdb redis.UniversalClient,
) Service {
	watermillLogger := watermill.NewSlogLogger(log.FromContext(context.Background()))
	publisher := ticketsMessage.NewRedisPublisher(rdb, watermillLogger)
	eventBus := ticketsEvent.NewEventBus(publisher, watermillLogger)
	redisSubscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        rdb,
			ConsumerGroup: "events_splitter",
		}, watermillLogger,
	)
	if err != nil {
		panic(err)
	}

	redisSubscriberStore, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        rdb,
			ConsumerGroup: "events_store",
		}, watermillLogger,
	)
	if err != nil {
		panic(err)
	}

	ticketRepo := ticketsDB.NewTicketsRepository(dbConn)
	showRepo := ticketsDB.NewShowsRepository(dbConn)
	bookingRepo := ticketsDB.NewBookingRepository(dbConn)
	eventRepo := ticketsDB.NewEventsRepository(dbConn)

	commandBus := ticketsCommand.NewCommandBus(publisher, watermillLogger)
	commandProcessorConfig := ticketsCommand.NewCommandProcessorConfig(
		rdb,
		watermillLogger,
	)
	commandHandler := ticketsCommand.NewCommandHandler(
		receiptsService,
		paymentService,
		eventBus,
	)

	postgresSubscriber := ticketsOutbox.NewPostgresSubscriber(dbConn, watermillLogger)

	eventHandler := ticketsEvent.NewEventHandler(
		spreadsheetsAPI,
		receiptsService,
		fileService,
		deadNationService,
		ticketRepo,
		showRepo,
		eventRepo,
		eventBus,
	)
	eventProcessorConfig := ticketsEvent.NewEventProcessorConfig(
		rdb,
		watermillLogger,
	)
	opsReadModel := ticketsDB.NewOpsBookingReadModel(dbConn, eventBus)
	router := ticketsMessage.NewRouter(
		redisSubscriberStore,
		redisSubscriber,
		postgresSubscriber,
		publisher,
		*eventProcessorConfig,
		*commandProcessorConfig,
		*commandHandler,
		opsReadModel,
		eventHandler,
		watermillLogger,
	)

	echoRouter := ticketsHttp.NewHttpRouter(
		eventBus,
		commandBus,
		ticketRepo,
		showRepo,
		bookingRepo,
		opsReadModel,
	)

	return Service{
		db:            dbConn,
		echoRouter:    echoRouter,
		messageRouter: router,
		opsReadModel:  opsReadModel,
		eventRepo:     eventRepo,
	}
}

func (s Service) Run(ctx context.Context) error {
	if err := ticketsDB.InitializeDatabaseSchema(s.db); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}
	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.Go(
		func() error {
			return s.messageRouter.Run(ctx)
		},
	)
	errGroup.Go(
		func() error {
			<-s.messageRouter.Running()

			err := s.echoRouter.Start(":8080")
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	)

	go func() {
		if err := readModelMigration.MigrateReadModel(ctx, s.eventRepo, s.opsReadModel); err != nil {
			log.FromContext(ctx).With("error", err).Error("failed to migrate read model")
		}
	}()

	errGroup.Go(
		func() error {
			<-ctx.Done()
			return s.echoRouter.Shutdown(context.Background())
		},
	)

	return errGroup.Wait()
}
