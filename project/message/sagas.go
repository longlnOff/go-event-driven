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
	vb, err := v.vipBundleRepository.Get(ctx, ticketsEntity.VipBundleID{vipBundleID})
	if err != nil {
		return err
	}
	ticketIDs := vb.TicketIDs
	if len(ticketIDs) < vb.NumberOfTickets {
		return errors.New("confirmed tickets haven't processed yet")
	}
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
