package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func health(c echo.Context) error {
	go func() {
		for {
			veryImportantCounter.Inc()
			time.Sleep(time.Millisecond * 100)
		}
	}()

	return c.String(http.StatusOK, "ok")
}

var (
	veryImportantCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			// metric will be named tickets_very_important_counter_total
			Namespace: "tickets",
			Name:      "very_important_counter_total",
			Help:      "Total number of very important things processed",
		},
	)
)
