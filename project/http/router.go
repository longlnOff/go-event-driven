package http

import (
	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewHttpRouter(
	eventBus *cqrs.EventBus,
	commandBus *cqrs.CommandBus,
	ticketRepo TicketsRepository,
	showRepo ShowsRepository,
	bookingRepo BookingRepository,
	opsReadModel OpsBookingReadModel,
	vipBundleRepo VipBundleRepository,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		eventBus:          eventBus,
		commandBus:        commandBus,
		ticketRepository:  ticketRepo,
		showRepository:    showRepo,
		bookingRepository: bookingRepo,
		opsReadModel:      opsReadModel,
		vipBundleRepo:     vipBundleRepo,
	}

	e.GET("/health", health)
	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/tickets", handler.GetAllTickets)

	e.POST("/shows", handler.CreateShow)

	e.POST("/book-tickets", handler.CreateBooking)

	e.PUT("ticket-refund/:ticket_id", handler.PutTicketRefund)
	e.GET("/ops/bookings", handler.GetAllBookingByDate)
	e.GET("/ops/bookings/:id", handler.GetBookingByID)

	// vip bundle
	e.POST("/book-vip-bundle", handler.PostVipBundle)

	// for metrics
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// use this for add spans for all coming http request
	e.Use(otelecho.Middleware("tickets"))

	return e
}
