package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/transportation"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/samber/lo"

	"tickets/entities"
)

var ErrNoFlightTicketsAvailable = fmt.Errorf("no flight tickets available")

type TransportationClient struct {
	clients *clients.Clients
}

func NewTransportationClient(clients *clients.Clients) *TransportationClient {
	if clients == nil {
		panic("clients is nil")
	}

	return &TransportationClient{clients: clients}
}

func (t TransportationClient) BookFlight(
	ctx context.Context,
	request entities.BookFlightTicketRequest,
) (entities.BookFlightTicketResponse, error) {
	resp, err := t.clients.Transportation.PutFlightTicketsWithResponse(
		ctx, transportation.BookFlightTicketRequest{
			CustomerEmail:  request.CustomerEmail,
			FlightId:       request.FlightID,
			PassengerNames: request.PassengerNames,
			ReferenceId:    request.ReferenceId,
			IdempotencyKey: request.IdempotencyKey,
		},
	)
	if err != nil {
		return entities.BookFlightTicketResponse{}, fmt.Errorf("failed to book flight: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusConflict:
		return entities.BookFlightTicketResponse{}, ErrNoFlightTicketsAvailable
	case http.StatusCreated:
		return entities.BookFlightTicketResponse{
			TicketIds: lo.Map(
				resp.JSON201.TicketIds, func(i openapi_types.UUID, _ int) uuid.UUID {
					return i
				},
			),
		}, nil
	default:
		return entities.BookFlightTicketResponse{}, fmt.Errorf(
			"unexpected status code for PUT transportation-api/transportation/flight-tickets: %d",
			resp.StatusCode(),
		)
	}
}
