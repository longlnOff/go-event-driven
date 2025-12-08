package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(
	publisher message.Publisher,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		publisher: publisher,
	}

	e.GET("/health", func(c echo.Context) error {return c.String(http.StatusOK, "ok")})
	e.POST("/tickets-confirmation", handler.PostTicketsConfirmation)


	return e
}
