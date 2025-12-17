package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	ticketsEntity "tickets/entities"
)

type TicketStatusRequest struct {
	TicketID      string              `json:"ticket_id"`
	Status        string              `json:"status"`
	Price         ticketsEntity.Money `json:"price"`
	CustomerEmail string              `json:"customer_email"`
}

type TicketsStatusRequest struct {
	Tickets []TicketStatusRequest `json:"tickets"`
}

func (h Handler) PostTicketsStatus(c echo.Context) error {
	var request TicketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for i := range request.Tickets {
		ticket := request.Tickets[i]

		if ticket.Status == "confirmed" {
			event := ticketsEntity.TicketBookingConfirmed{
				Header:        ticketsEntity.NewMessageHeader(),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}
			err = h.eventBus.Publish(
				c.Request().Context(),
				event,
			)
			if err != nil {
				return fmt.Errorf("failed to publish event TicketBookingConfirmed: %w", err)
			}
		} else if ticket.Status == "canceled" {
			event := ticketsEntity.TicketBookingCanceled{
				Header:        ticketsEntity.NewMessageHeader(),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}
			err = h.eventBus.Publish(
				c.Request().Context(),
				event,
			)
			if err != nil {
				return fmt.Errorf("failed to publish event TicketBookingCanceled: %w", err)
			}
		} else {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}
