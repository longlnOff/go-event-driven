package http

import (
	"errors"
	"net/http"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/labstack/echo/v4"
)

type CreateBookingRequest struct {
	ShowID          string `json:"show_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	CustomerEmail   string `json:"customer_email"`
}

type BookingResponse struct {
	BookingID string `json:"booking_id"`
}

func (h Handler) CreateBooking(c echo.Context) error {
	var request CreateBookingRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	if request.NumberOfTickets < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "number of tickets must be greater than 0")
	}
	booking := ticketsEntity.Booking{
		BookingID:       watermill.NewUUID(),
		ShowID:          request.ShowID,
		NumberOfTickets: request.NumberOfTickets,
		CustomerEmail:   request.CustomerEmail,
	}
	err = h.bookingRepository.AddBooking(c.Request().Context(), booking)
	if err != nil {
		if errors.As(err, &echo.ErrBadRequest) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, BookingResponse{BookingID: booking.BookingID})
}
