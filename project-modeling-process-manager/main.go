package main

import (
	"context"
	"errors"
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
	ShowId          uuid.UUID  `json:"show_id"`
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

func NewVipBundle(
	vipBundleID VipBundleID,
	bookingID uuid.UUID,
	customerEmail string,
	numberOfTickets int,
	showId uuid.UUID,
	passengers []string,
	inboundFlightID uuid.UUID,
	returnFlightID uuid.UUID,
) (*VipBundle, error) {
	if vipBundleID.UUID == uuid.Nil {
		return nil, fmt.Errorf("vip bundle id must be set")
	}
	if bookingID == uuid.Nil {
		return nil, fmt.Errorf("booking id must be set")
	}
	if customerEmail == "" {
		return nil, fmt.Errorf("customer email must be set")
	}
	if numberOfTickets <= 0 {
		return nil, fmt.Errorf("number of tickets must be greater than 0")
	}
	if showId == uuid.Nil {
		return nil, fmt.Errorf("show id must be set")
	}
	if numberOfTickets != len(passengers) {
		return nil, fmt.Errorf("number of tickets and passengers count mismatch")
	}
	if inboundFlightID == uuid.Nil {
		return nil, fmt.Errorf("inbound flight id must be set")
	}
	if returnFlightID == uuid.Nil {
		return nil, fmt.Errorf("return flight id must be set")
	}

	return &VipBundle{
		VipBundleID:     vipBundleID,
		BookingID:       bookingID,
		CustomerEmail:   customerEmail,
		NumberOfTickets: numberOfTickets,
		ShowId:          showId,
		Passengers:      passengers,
		InboundFlightID: inboundFlightID,
		ReturnFlightID:  returnFlightID,
	}, nil
}

type VipBundleRepository interface {
	Add(ctx context.Context, vipBundle VipBundle) error
	Get(ctx context.Context, vipBundleID VipBundleID) (VipBundle, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (VipBundle, error)

	UpdateByID(
		ctx context.Context,
		vipBundleID VipBundleID,
		updateFn func(vipBundle VipBundle) (VipBundle, error),
	) (VipBundle, error)

	UpdateByBookingID(
		ctx context.Context,
		bookingID uuid.UUID,
		updateFn func(vipBundle VipBundle) (VipBundle, error),
	) (VipBundle, error)
}

type VipBundleProcessManager struct {
	commandBus CommandBus
	eventBus   EventBus
	repository VipBundleRepository
}

func NewVipBundleProcessManager(
	commandBus CommandBus,
	eventBus EventBus,
	repository VipBundleRepository,
) *VipBundleProcessManager {
	return &VipBundleProcessManager{
		commandBus: commandBus,
		eventBus:   eventBus,
		repository: repository,
	}
}

func (v VipBundleProcessManager) OnVipBundleInitialized(ctx context.Context, event *VipBundleInitialized_v1) error {
	// Get information about VIP bundle
	vipBundle, err := v.repository.Get(ctx, event.VipBundleID)
	if err != nil {
		return err
	}
	bookingShowTickets := BookShowTickets{
		BookingID:       vipBundle.BookingID,
		CustomerEmail:   vipBundle.CustomerEmail,
		NumberOfTickets: vipBundle.NumberOfTickets,
		ShowId:          vipBundle.ShowId,
	}
	// publish command
	return v.commandBus.Send(ctx, bookingShowTickets)
}

func (v VipBundleProcessManager) OnBookingMade(ctx context.Context, event *BookingMade_v1) error {
	vipBundle, err := v.repository.UpdateByBookingID(
		ctx,
		event.BookingID,
		func(b VipBundle) (VipBundle, error) {
			b.BookingMadeAt = &event.Header.PublishedAt
			return b, nil
		},
	)
	if err != nil {
		return err
	}

	return v.commandBus.Send(
		ctx, BookFlight{
			CustomerEmail:  vipBundle.CustomerEmail,
			FlightID:       vipBundle.InboundFlightID,
			Passengers:     vipBundle.Passengers,
			ReferenceID:    vipBundle.VipBundleID.String(),
			IdempotencyKey: uuid.NewString(),
		},
	)
}

func (v VipBundleProcessManager) OnFlightBooked(ctx context.Context, event *FlightBooked_v1) error {
	vb, err := v.repository.UpdateByID(
		ctx,
		MustParseBundleID(event.ReferenceID),
		func(vipBundle VipBundle) (VipBundle, error) {
			// Check if this is inbound or return flight
			if event.FlightID == vipBundle.InboundFlightID {
				// Store inbound flight ticket IDs
				vipBundle.InboundFlightTicketsIDs = event.TicketIDs
			} else if event.FlightID == vipBundle.ReturnFlightID {
				// Store return flight ticket IDs and timestamp
				vipBundle.ReturnFlightTicketsIDs = event.TicketIDs
				vipBundle.ReturnFlightBookedAt = &event.Header.PublishedAt
			}
			return vipBundle, nil
		},
	)
	if err != nil {
		return err
	}

	// Determine next action based on which flight was booked
	if event.FlightID == vb.InboundFlightID {
		// Inbound flight booked - book return flight
		return v.commandBus.Send(
			ctx, BookFlight{
				CustomerEmail:  vb.CustomerEmail,
				FlightID:       vb.ReturnFlightID,
				Passengers:     vb.Passengers,
				ReferenceID:    vb.VipBundleID.String(),
				IdempotencyKey: uuid.NewString(),
			},
		)
	} else if event.FlightID == vb.ReturnFlightID {
		// Return flight booked - book taxi
		return v.commandBus.Send(
			ctx, BookTaxi{
				CustomerEmail:      vb.CustomerEmail,
				CustomerName:       vb.Passengers[0],
				NumberOfPassengers: vb.NumberOfTickets,
				ReferenceID:        vb.VipBundleID.String(),
				IdempotencyKey:     uuid.NewString(),
			},
		)
	}

	return nil
}

func (v VipBundleProcessManager) OnBookingFailed(ctx context.Context, event *BookingFailed_v1) error {
	vb, err := v.repository.UpdateByBookingID(
		ctx,
		event.BookingID,
		func(vipBundle VipBundle) (VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return err
	}
	return v.eventBus.Publish(
		ctx, VipBundleFinalized_v1{
			Header:      NewMessageHeader(),
			VipBundleID: vb.VipBundleID,
			Success:     false,
		},
	)
}

func (v VipBundleProcessManager) OnTicketBookingConfirmed(ctx context.Context, event *TicketBookingConfirmed_v1) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return err
	}
	_, err = v.repository.UpdateByBookingID(
		ctx,
		bookingID,
		func(vipBundle VipBundle) (VipBundle, error) {
			ticketID, err := uuid.Parse(event.TicketID)
			if err != nil {
				return VipBundle{}, err
			}
			vipBundle.TicketIDs = append(vipBundle.TicketIDs, ticketID)
			return vipBundle, nil
		},
	)
	return err
}

func (v VipBundleProcessManager) OnFlightBookingFailed(ctx context.Context, event *FlightBookingFailed_v1) error {
	vipBundleID, err := uuid.Parse(event.ReferenceID)
	if err != nil {
		return err
	}
	vb, err := v.repository.Get(ctx, VipBundleID{vipBundleID})
	if err != nil {
		return err
	}
	ticketIDs := vb.TicketIDs
	if len(ticketIDs) < vb.NumberOfTickets {
		return errors.New("confirmed tickets haven't processed yet")
	}

	// Refund show tickets
	for i := range ticketIDs {
		err := v.commandBus.Send(
			ctx, RefundTicket{
				Header:   NewMessageHeader(),
				TicketID: ticketIDs[i].String(),
			},
		)
		if err != nil {
			return err
		}
	}

	// Cancel inbound flight tickets if they exist
	if len(vb.InboundFlightTicketsIDs) > 0 {
		err := v.commandBus.Send(
			ctx, CancelFlightTickets{
				FlightTicketIDs: vb.InboundFlightTicketsIDs,
			},
		)
		if err != nil {
			return err
		}
	}

	vb, err = v.repository.UpdateByBookingID(
		ctx,
		vb.BookingID,
		func(vipBundle VipBundle) (VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return err
	}

	return v.eventBus.Publish(
		ctx, VipBundleFinalized_v1{
			Header:      NewMessageHeader(),
			VipBundleID: vb.VipBundleID,
			Success:     false,
		},
	)
}

func (v VipBundleProcessManager) OnTaxiBooked(ctx context.Context, event *TaxiBooked_v1) error {
	vb, err := v.repository.UpdateByID(
		ctx,
		MustParseBundleID(event.ReferenceID),
		func(vipBundle VipBundle) (VipBundle, error) {
			vipBundle.TaxiBookedAt = &event.Header.PublishedAt
			vipBundle.TaxiBookingID = &event.TaxiBookingID
			vipBundle.IsFinalized = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return err
	}

	return v.eventBus.Publish(
		ctx, VipBundleFinalized_v1{
			Header:      NewMessageHeader(),
			VipBundleID: vb.VipBundleID,
			Success:     true,
		},
	)
}

func (v VipBundleProcessManager) OnTaxiBookingFailed(ctx context.Context, event *TaxiBookingFailed_v1) error {
	vipBundleID, err := uuid.Parse(event.ReferenceID)
	if err != nil {
		return err
	}
	vb, err := v.repository.Get(ctx, VipBundleID{vipBundleID})
	if err != nil {
		return err
	}

	// Refund show tickets
	for i := range vb.TicketIDs {
		err := v.commandBus.Send(
			ctx, RefundTicket{
				Header:   NewMessageHeader(),
				TicketID: vb.TicketIDs[i].String(),
			},
		)
		if err != nil {
			return err
		}
	}

	// Cancel inbound flight tickets
	if len(vb.InboundFlightTicketsIDs) > 0 {
		err := v.commandBus.Send(
			ctx, CancelFlightTickets{
				FlightTicketIDs: vb.InboundFlightTicketsIDs,
			},
		)
		if err != nil {
			return err
		}
	}

	// Cancel return flight tickets
	if len(vb.ReturnFlightTicketsIDs) > 0 {
		err := v.commandBus.Send(
			ctx, CancelFlightTickets{
				FlightTicketIDs: vb.ReturnFlightTicketsIDs,
			},
		)
		if err != nil {
			return err
		}
	}

	vb, err = v.repository.UpdateByBookingID(
		ctx,
		vb.BookingID,
		func(vipBundle VipBundle) (VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return err
	}

	return v.eventBus.Publish(
		ctx, VipBundleFinalized_v1{
			Header:      NewMessageHeader(),
			VipBundleID: vb.VipBundleID,
			Success:     false,
		},
	)
}
