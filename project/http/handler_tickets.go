package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	ticketsEntity "tickets/entities"
)

type TicketStatusRequest struct {
	BookingId     string              `json:"booking_id"`
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
	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Idempotency-Key is required")
	}
	for i := range request.Tickets {
		ticket := request.Tickets[i]
		key := idempotencyKey + ticket.TicketID
		if ticket.Status == "confirmed" {
			event := ticketsEntity.TicketBookingConfirmed{
				Header:        ticketsEntity.NewMessageHeaderWithIdempotencyKey(key),
				BookingID:     ticket.BookingId,
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
				Header:        ticketsEntity.NewMessageHeaderWithIdempotencyKey(key),
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

type TicketResponse struct {
	TicketID      string              `json:"ticket_id"`
	CustomerEmail string              `json:"customer_email"`
	Price         ticketsEntity.Money `json:"price"`
}

func (h Handler) GetAllTickets(c echo.Context) error {
	tickets, err := h.ticketRepository.FindAll(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get tickets: %w", err)
	}

	var response []TicketResponse
	for i := range tickets {
		ticket := tickets[i]
		response = append(response, TicketResponse{
			TicketID:      ticket.TicketID,
			CustomerEmail: ticket.CustomerEmail,
			Price:         ticket.Price,
		})
	}

	return c.JSON(http.StatusOK, response)
}
