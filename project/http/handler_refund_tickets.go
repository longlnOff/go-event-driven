package http

import (
	"net/http"
	ticketsEntity "tickets/entities"

	"github.com/labstack/echo/v4"
)

type TicketRefundResponse struct {
	TicketID string `json:"ticket_id"`
}

func (h Handler) PutTicketRefund(c echo.Context) error {
	ticketID := c.Param("ticket_id")
	if ticketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_id is required")
	}

	refundTicket := ticketsEntity.RefundTicket{
		Header:   ticketsEntity.NewMessageHeader(),
		TicketID: ticketID,
	}
	err := h.commandBus.Send(c.Request().Context(), &refundTicket)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusAccepted, TicketRefundResponse{TicketID: refundTicket.TicketID})
}
