package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func health(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
