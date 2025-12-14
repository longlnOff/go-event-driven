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

	correlationID := c.Request().Header.Get("Correlation-ID")

	if err != nil {
		return err
	}

	for i := range request.Tickets {
		ticket := request.Tickets[i]
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
		msg := message.NewMessage(watermill.NewUUID(), []byte(messageData))
		msg.Metadata.Set("correlation_id", correlationID)

		if ticket.Status == "confirmed" {
			msg.Metadata.Set("type", ticketsEvent.TicketBookingConfirmedTopic)
			err = h.publisher.Publish(
				ticketsEvent.TicketBookingConfirmedTopic,
				msg,
			)
			if err != nil {
				return err
			}
		} else if ticket.Status == "canceled" {
			msg.Metadata.Set("type", ticketsEvent.TicketBookingCanceledTopic)
			err = h.publisher.Publish(
				ticketsEvent.TicketBookingCanceledTopic,
				msg,
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
