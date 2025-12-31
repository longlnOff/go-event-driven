package event

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) StoreTickets(
	ctx context.Context,
	event *ticketsEntity.TicketBookingConfirmed_v1,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Storing ticket")
	err := h.ticketRepository.Add(
		ctx,
		ticketsEntity.Ticket{
			TicketID: event.TicketID,
			Price: ticketsEntity.Money{
				Amount:   event.Price.Amount,
				Currency: event.Price.Currency,
			},
			CustomerEmail: event.CustomerEmail,
		},
	)

	return err
}

func (h Handler) RemoveCanceledTicket(
	ctx context.Context,
	event *ticketsEntity.TicketBookingCanceled_v1,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Removing ticket")
	err := h.ticketRepository.Remove(
		ctx,
		ticketsEntity.Ticket{
			TicketID: event.TicketID,
			Price: ticketsEntity.Money{
				Amount:   event.Price.Amount,
				Currency: event.Price.Currency,
			},
			CustomerEmail: event.CustomerEmail,
		},
	)

	return err
}
