package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(
	eventBus *cqrs.EventBus,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		eventBus: eventBus,
	}

	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })
	e.POST("/tickets-status", handler.PostTicketsStatus)

	return e
}
