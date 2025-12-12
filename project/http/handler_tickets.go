package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"

	ticketsEntity "tickets/entities"
	ticketsEvent "tickets/message/event"
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
			messageData, err := json.Marshal(event)
			if err != nil {
				return err
			}

			err = h.publisher.Publish(
				ticketsEvent.TicketBookingConfirmedTopic,
				message.NewMessage(watermill.NewUUID(), []byte(messageData)),
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
