package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

func (h Handler) GetAllBookingByDate(c echo.Context) error {
	receiptIssueDate := c.QueryParam("receipt_issue_date")
	allBooking, err := h.opsReadModel.AllBookingsByDate(receiptIssueDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, allBooking)
}

func (h Handler) GetBookingByID(c echo.Context) error {
	bookingID := c.Param("id")
	if bookingID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "booking_id is required")
	}

	booking, err := h.opsReadModel.ReservationReadModel(
		c.Request().Context(),
		bookingID,
	)

	if err != nil {

		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Code.Name() == "unique_violation" {
			// handling re-delivery
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		// return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		return echo.NewHTTPError(http.StatusNotFound, err.Error())

	}
	return c.JSON(http.StatusOK, booking)
}
