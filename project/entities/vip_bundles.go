package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type VipBundleID struct {
	uuid.UUID
}

func ParseBundleID(id string) (VipBundleID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return VipBundleID{}, fmt.Errorf("failed to parse VipBundleID: %w", err)
	}
	return VipBundleID{UUID: parsed}, nil
}

func MustParseBundleID(id string) VipBundleID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		panic(fmt.Sprintf("failed to parse VipBundleID: %s", id))
	}
	return VipBundleID{UUID: parsed}
}

type VipBundle struct {
	VipBundleID VipBundleID `json:"vip_bundle_id"`

	BookingID       uuid.UUID  `json:"booking_id"`
	CustomerEmail   string     `json:"customer_email"`
	NumberOfTickets int        `json:"number_of_tickets"`
	ShowID          uuid.UUID  `json:"show_id"`
	BookingMadeAt   *time.Time `json:"booking_made_at"`

	TicketIDs []uuid.UUID `json:"ticket_ids"`

	Passengers []string `json:"passengers"`

	InboundFlightID uuid.UUID `json:"inbound_flight_id"`
	IsFinalized     bool      `json:"is_finalized"`
	Failed          bool      `json:"failed"`

	InboundFlightTicketsIDs []uuid.UUID `json:"inbound_flight_tickets_ids"`
	ReturnFlightID          uuid.UUID   `json:"return_flight_id"`

	ReturnFlightBookedAt   *time.Time  `json:"return_flight_booked_at"`
	ReturnFlightTicketsIDs []uuid.UUID `json:"return_flight_tickets_ids"`

	TaxiBookedAt  *time.Time `json:"taxi_booked_at"`
	TaxiBookingID *uuid.UUID `json:"taxi_booking_id"`
}
