package entities

import "github.com/google/uuid"

type BookFlightTicketRequest struct {
	CustomerEmail  string
	FlightID       uuid.UUID
	PassengerNames []string
	ReferenceId    string
	IdempotencyKey string
}

type BookFlightTicketResponse struct {
	TicketIds []uuid.UUID `json:"ticket_ids"`
}
