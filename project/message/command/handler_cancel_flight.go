package command

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) CancelFlight(
	ctx context.Context,
	command *ticketsEntity.CancelFlightTickets,
) error {
	logger := log.FromContext(ctx)
	logger.Info("CancelFlight start")
	return h.bookFlightService.CancelFlightTickets(
		ctx,
		ticketsEntity.CancelFlightTicketsRequest{
			TicketIds: command.FlightTicketIDs,
		},
	)

}
