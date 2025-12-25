package http

import (
	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(
	eventBus *cqrs.EventBus,
	ticketRepo TicketsRepository,
	showRepo ShowsRepository,
	bookingRepo BookingRepository,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		eventBus:          eventBus,
		ticketRepository:  ticketRepo,
		showRepository:    showRepo,
		bookingRepository: bookingRepo,
	}

	e.GET("/health", health)
	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/tickets", handler.GetAllTickets)

	e.POST("/shows", handler.CreateShow)

	e.POST("/book-tickets", handler.CreateBooking)

	return e
}
