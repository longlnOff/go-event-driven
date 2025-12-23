package adapters

import (
	"context"
	"fmt"
	"net/http"
	ticketsEntity "tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/dead_nation"
)

type DeadNationServiceClient struct {
	// we are not mocking this client: it's pointless to use interface here
	clients *clients.Clients
}

func NewDeadNationServiceClient(clients *clients.Clients) *DeadNationServiceClient {
	if clients == nil {
		panic("NewDeadNationServiceClient: clients is nil")
	}

	return &DeadNationServiceClient{clients: clients}
}

func (c DeadNationServiceClient) CallDeadNation(
	ctx context.Context,
	booking ticketsEntity.DeadNationBooking,
) error {
	resp, err := c.clients.DeadNation.PostTicketBookingWithResponse(
		ctx,
		dead_nation.PostTicketBookingRequest{
			BookingId:       booking.BookingID,
			EventId:         booking.DeadNationEventID,
			NumberOfTickets: booking.NumberOfTickets,
			CustomerAddress: booking.CustomerEmail,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to call Dead Nation: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		// receipt already exists
		return nil
	default:
		return fmt.Errorf("unexpected status code for call Dead Nation: %d", resp.StatusCode())
	}
}
