package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	ticketsEntity "tickets/entities"
)

type vipBundleRequest struct {
	CustomerEmail   string    `json:"customer_email"`
	InboundFlightID uuid.UUID `json:"inbound_flight_id"`
	NumberOfTickets int       `json:"number_of_tickets"`
	Passengers      []string  `json:"passengers"`
	ShowID          uuid.UUID `json:"show_id"`
}

type vipBundleResponse struct {
	BookingID   uuid.UUID                 `json:"booking_id"`
	VipBundleID ticketsEntity.VipBundleID `json:"vip_bundle_id"`
}

func (h Handler) PostVipBundle(c echo.Context) error {
	var request vipBundleRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	if request.NumberOfTickets < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "number of tickets must be greater than 0")
	}

	vb := ticketsEntity.VipBundle{
		VipBundleID:     ticketsEntity.VipBundleID{uuid.New()},
		BookingID:       uuid.New(),
		CustomerEmail:   request.CustomerEmail,
		NumberOfTickets: request.NumberOfTickets,
		ShowID:          request.ShowID,
		Passengers:      request.Passengers,
		InboundFlightID: request.InboundFlightID,
		IsFinalized:     false,
		Failed:          false,
	}

	if err := h.vipBundleRepo.Add(c.Request().Context(), vb); err != nil {
		return err
	}

	return c.JSON(
		http.StatusCreated, vipBundleResponse{
			BookingID:   vb.BookingID,
			VipBundleID: vb.VipBundleID,
		},
	)
}
