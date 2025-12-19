package event

import (
	"context"
	"fmt"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) PrintTickets(
	ctx context.Context,
	event *ticketsEntity.TicketBookingConfirmed,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Printing ticket")
	body := `
		<html>
			<head>
				<title>Ticket</title>
			</head>
			<body>
				<h1>Ticket ` + event.TicketID + `</h1>
				<p>Price: ` + event.Price.Amount + ` ` + event.Price.Currency + `</p>	
			</body>
		</html>`
	fileName := fmt.Sprintf("%s-ticket.html", event.TicketID)

	err := h.fileService.UpLoadFile(ctx, fileName, body)
	if err != nil {
		return err
	}

	ticketPrintedEvent := ticketsEntity.TicketPrinted{
		Header:   ticketsEntity.NewMessageHeaderWithIdempotencyKey(event.Header.IdempotencyKey),
		TicketID: event.TicketID,
		FileName: fileName,
	}
	err = h.eventBus.Publish(ctx, ticketPrintedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}
