package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
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

			eventReceipt := ticketsEntity.IssueReceiptPayload{
				TicketID: ticket.TicketID,
				Price:    ticket.Price,
			}
			messageReceipt, err := json.Marshal(eventReceipt)
			if err != nil {
				return err
			}

			err = h.publisher.Publish(
				"issue-receipt",
				message.NewMessage(watermill.NewUUID(), []byte(messageReceipt)),
			)
			if err != nil {
				return err
			}

			eventTracker := ticketsEntity.AppendToTrackerPayload{
				TicketID:      ticket.TicketID,
				Price:         ticket.Price,
				CustomerEmail: ticket.CustomerEmail,
			}
			messageTracker, err := json.Marshal(eventTracker)
			if err != nil {
				return err
			}
			err = h.publisher.Publish(
				"append-to-tracker",
				message.NewMessage(watermill.NewUUID(), []byte(messageTracker)),
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
