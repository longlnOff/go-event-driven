package message

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ticketsEntity "tickets/entities"

	"github.com/google/uuid"
)

func MustParseBundleID(id string) ticketsEntity.VipBundleID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		panic(fmt.Sprintf("failed to parse VipBundleID: %s", id))
	}
	return ticketsEntity.VipBundleID{UUID: parsed}
}

type VipBundleRepository interface {
	Add(ctx context.Context, vipBundle ticketsEntity.VipBundle) error
	Get(ctx context.Context, vipBundleID ticketsEntity.VipBundleID) (ticketsEntity.VipBundle, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (ticketsEntity.VipBundle, error)

	UpdateByID(
		ctx context.Context,
		vipBundleID ticketsEntity.VipBundleID,
		updateFn func(vipBundle ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error),
	) (ticketsEntity.VipBundle, error)

	UpdateByBookingID(
		ctx context.Context,
		bookingID uuid.UUID,
		updateFn func(vipBundle ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error),
	) (ticketsEntity.VipBundle, error)
}

type CommandBus interface {
	Send(ctx context.Context, command any) error
}

type EventBus interface {
	Publish(ctx context.Context, event any) error
}

type VipBundleProcessManager struct {
	commandBus          CommandBus
	eventBus            EventBus
	vipBundleRepository VipBundleRepository
}

func NewVipBundleProcessManager(
	commandBus CommandBus,
	eventBus EventBus,
	vipBundleRepository VipBundleRepository,
) *VipBundleProcessManager {
	return &VipBundleProcessManager{
		commandBus:          commandBus,
		eventBus:            eventBus,
		vipBundleRepository: vipBundleRepository,
	}
}

func (v VipBundleProcessManager) OnVipBundleInitialized(
	ctx context.Context,
	event *ticketsEntity.VipBundleInitialized_v1,
) error {
	// Get information about VIP bundle
	vipBundle, err := v.vipBundleRepository.Get(ctx, event.VipBundleID)
	if err != nil {
		return err
	}
	bookingShowTickets := ticketsEntity.BookShowTickets{
		BookingID:       vipBundle.BookingID,
		CustomerEmail:   vipBundle.CustomerEmail,
		NumberOfTickets: vipBundle.NumberOfTickets,
		ShowId:          vipBundle.ShowID,
	}
	// publish command
	return v.commandBus.Send(ctx, bookingShowTickets)
}

func (v VipBundleProcessManager) OnBookingMade(ctx context.Context, event *ticketsEntity.BookingMade_v1) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return err
	}
	vipBundle, err := v.vipBundleRepository.UpdateByBookingID(
		ctx,
		bookingID,
		func(b ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error) {
			b.BookingMadeAt = &event.Header.PublishedAt
			return b, nil
		},
	)
	if err != nil {
		// If the booking is not part of a VIP bundle, ignore the event
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return v.commandBus.Send(
		ctx, ticketsEntity.BookFlight{
			CustomerEmail:  vipBundle.CustomerEmail,
			FlightID:       vipBundle.InboundFlightID,
			Passengers:     vipBundle.Passengers,
			ReferenceID:    vipBundle.VipBundleID.String(),
			IdempotencyKey: uuid.NewString(),
		},
	)
}

func (v VipBundleProcessManager) OnFlightBooked(ctx context.Context, event *ticketsEntity.FlightBooked_v1) error {
	vb, err := v.vipBundleRepository.UpdateByID(
		ctx,
		MustParseBundleID(event.ReferenceID),
		func(vipBundle ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error) {
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
			ctx, ticketsEntity.BookFlight{
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
			ctx, ticketsEntity.BookTaxi{
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

func (v VipBundleProcessManager) OnBookingFailed(ctx context.Context, event *ticketsEntity.BookingFailed_v1) error {
	vb, err := v.vipBundleRepository.UpdateByBookingID(
		ctx,
		event.BookingID,
		func(vipBundle ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		// If the booking is not part of a VIP bundle, ignore the event
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return v.eventBus.Publish(
		ctx, ticketsEntity.VipBundleFinalized_v1{
			Header:      ticketsEntity.NewMessageHeader(),
			VipBundleID: vb.VipBundleID,
			Success:     false,
		},
	)
}

func (v VipBundleProcessManager) OnTicketBookingConfirmed(
	ctx context.Context,
	event *ticketsEntity.TicketBookingConfirmed_v1,
) error {
	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return err
	}
	_, err = v.vipBundleRepository.UpdateByBookingID(
		ctx,
		bookingID,
		func(vipBundle ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error) {
			ticketID, err := uuid.Parse(event.TicketID)
			if err != nil {
				return ticketsEntity.VipBundle{}, err
			}
			vipBundle.TicketIDs = append(vipBundle.TicketIDs, ticketID)
			return vipBundle, nil
		},
	)
	// If the booking is not part of a VIP bundle, ignore the event
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func (v VipBundleProcessManager) OnFlightBookingFailed(
	ctx context.Context,
	event *ticketsEntity.FlightBookingFailed_v1,
) error {
	vipBundleID, err := uuid.Parse(event.ReferenceID)
	if err != nil {
		return err
	}
	vb, err := v.vipBundleRepository.Get(ctx, ticketsEntity.VipBundleID{UUID: vipBundleID})
	if err != nil {
		return err
	}
	ticketIDs := vb.TicketIDs
	if len(ticketIDs) < vb.NumberOfTickets {
		return errors.New("confirmed tickets haven't processed yet")
	}

	// refund shows
	for i := range ticketIDs {
		err := v.commandBus.Send(
			ctx, ticketsEntity.RefundTicket{
				Header:   ticketsEntity.NewMessageHeader(),
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
			ctx, ticketsEntity.CancelFlightTickets{
				FlightTicketIDs: vb.InboundFlightTicketsIDs,
			},
		)
		if err != nil {
			return err
		}
	}
	vb, err = v.vipBundleRepository.UpdateByBookingID(
		ctx,
		vb.BookingID,
		func(vipBundle ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return err
	}

	return v.eventBus.Publish(
		ctx, ticketsEntity.VipBundleFinalized_v1{
			Header:      ticketsEntity.NewMessageHeader(),
			VipBundleID: vb.VipBundleID,
			Success:     false,
		},
	)
}

func (v VipBundleProcessManager) OnTaxiBooked(ctx context.Context, event *ticketsEntity.TaxiBooked_v1) error {
	vb, err := v.vipBundleRepository.UpdateByID(
		ctx,
		MustParseBundleID(event.ReferenceID),
		func(vipBundle ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error) {
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
		ctx, ticketsEntity.VipBundleFinalized_v1{
			Header:      ticketsEntity.NewMessageHeader(),
			VipBundleID: vb.VipBundleID,
			Success:     true,
		},
	)
}

func (v VipBundleProcessManager) OnTaxiBookingFailed(
	ctx context.Context,
	event *ticketsEntity.TaxiBookingFailed_v1,
) error {
	vipBundleID, err := uuid.Parse(event.ReferenceID)
	if err != nil {
		return err
	}
	vb, err := v.vipBundleRepository.Get(ctx, ticketsEntity.VipBundleID{vipBundleID})
	if err != nil {
		return err
	}

	// Refund show tickets
	for i := range vb.TicketIDs {
		err := v.commandBus.Send(
			ctx, ticketsEntity.RefundTicket{
				Header:   ticketsEntity.NewMessageHeader(),
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
			ctx, ticketsEntity.CancelFlightTickets{
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
			ctx, ticketsEntity.CancelFlightTickets{
				FlightTicketIDs: vb.ReturnFlightTicketsIDs,
			},
		)
		if err != nil {
			return err
		}
	}

	vb, err = v.vipBundleRepository.UpdateByBookingID(
		ctx,
		vb.BookingID,
		func(vipBundle ticketsEntity.VipBundle) (ticketsEntity.VipBundle, error) {
			vipBundle.IsFinalized = true
			vipBundle.Failed = true
			return vipBundle, nil
		},
	)
	if err != nil {
		return err
	}

	return v.eventBus.Publish(
		ctx, ticketsEntity.VipBundleFinalized_v1{
			Header:      ticketsEntity.NewMessageHeader(),
			VipBundleID: vb.VipBundleID,
			Success:     false,
		},
	)
}
