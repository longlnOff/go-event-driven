package event

import (
	"context"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/google/uuid"
)

func (h Handler) CallDeadNation(
	ctx context.Context,
	event *ticketsEntity.BookingMade_v1,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Calling dead nation")

	show, err := h.showRepository.ShowByID(ctx, event.ShowID)
	if err != nil {
		return err
	}
	bookingID, _ := uuid.Parse(event.BookingID)
	deadNationID, _ := uuid.Parse(show.DeadNationID)
	deadNationBooking := ticketsEntity.DeadNationBooking{
		BookingID:         bookingID,
		NumberOfTickets:   event.NumberOfTickets,
		CustomerEmail:     event.CustomerEmail,
		DeadNationEventID: deadNationID,
	}
	return h.deadNationService.CallDeadNation(
		ctx,
		deadNationBooking,
	)
}
