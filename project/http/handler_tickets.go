package http

import (
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"

	ticketsEntity "tickets/entities"
)

type ticketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

func (h Handler) PostTicketsConfirmation(c echo.Context) error {
	var request ticketsConfirmationRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		err := h.publisher.Publish(
			"issue-receipt",
			message.NewMessage(watermill.NewUUID(), []byte(ticket)),
		)
		if err != nil {
			return err
		}

		err = h.publisher.Publish(
			"append-to-tracker",
			message.NewMessage(watermill.NewUUID(), []byte(ticket)),
		)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}



type TicketsStatusRequest struct {
	Tickets []TicketStatusRequest `json:"tickets"`
}

type TicketStatusRequest struct {
	TicketID      string `json:"ticket_id"`
	Status        string `json:"status"`
	Price         ticketsEntity.Money  `json:"price"`
	CustomerEmail string `json:"customer_email"`
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
			err := h.publisher.Publish(
				"issue-receipt",
				message.NewMessage(watermill.NewUUID(), []byte(ticket.TicketID)),
			)
			if err != nil {
				return err
			}

			err = h.publisher.Publish(
				"append-to-tracker",
				message.NewMessage(watermill.NewUUID(), []byte(ticket.TicketID)),
			)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}
