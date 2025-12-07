package http

import (
	"net/http"
	"github.com/labstack/echo/v4"

	ticketsWorker "tickets/worker"
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
		h.worker.Send(
			ticketsWorker.Message{
				Task: ticketsWorker.TaskIssueReceipt,
				TicketID: ticket,
			},
			ticketsWorker.Message{
				Task: ticketsWorker.TaskAppendToTracker,
				TicketID: ticket,
			},
		)
	}

	return c.NoContent(http.StatusOK)
}
