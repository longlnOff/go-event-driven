package http

import (
	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewHttpRouter(
	publisher message.Publisher,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		publisher: publisher,
	}

	e.POST("/tickets-confirmation", handler.PostTicketsConfirmation)

	return e
}
