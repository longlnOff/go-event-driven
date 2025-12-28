package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/shopspring/decimal"
)

type InvoiceIssued struct {
	InvoiceID    string
	CustomerName string
	Amount       decimal.Decimal
	IssuedAt     time.Time
}

type InvoicePaymentReceived struct {
	PaymentID  string
	InvoiceID  string
	PaidAmount decimal.Decimal
	PaidAt     time.Time

	FullyPaid bool
}

type InvoiceVoided struct {
	InvoiceID string
	VoidedAt  time.Time
}

type InvoiceReadModel struct {
	InvoiceID    string
	CustomerName string
	Amount       decimal.Decimal
	IssuedAt     time.Time

	FullyPaid     bool
	PaidAmount    decimal.Decimal
	LastPaymentAt time.Time

	Voided   bool
	VoidedAt time.Time
}

type InvoiceReadModelStorage struct {
	invoices map[string]InvoiceReadModel
	payments map[string]InvoicePaymentReceived
}

func NewInvoiceReadModelStorage() *InvoiceReadModelStorage {
	return &InvoiceReadModelStorage{
		invoices: make(map[string]InvoiceReadModel),
		payments: make(map[string]InvoicePaymentReceived),
	}
}

func (s *InvoiceReadModelStorage) Invoices() []InvoiceReadModel {
	invoices := make([]InvoiceReadModel, 0, len(s.invoices))
	for _, invoice := range s.invoices {
		invoices = append(invoices, invoice)
	}
	return invoices
}

func (s *InvoiceReadModelStorage) InvoiceByID(id string) (InvoiceReadModel, bool) {
	invoice, ok := s.invoices[id]
	return invoice, ok
}

func (s *InvoiceReadModelStorage) OnInvoiceIssued(ctx context.Context, event *InvoiceIssued) error {
	// TODO: implement
	_, ok := s.invoices[event.InvoiceID]
	if !ok {
		s.invoices[event.InvoiceID] = InvoiceReadModel{
			InvoiceID:    event.InvoiceID,
			CustomerName: event.CustomerName,
			IssuedAt:     event.IssuedAt,
			Amount:       event.Amount,
		}
	}

	return nil
}

func (s *InvoiceReadModelStorage) OnInvoicePaymentReceived(ctx context.Context, event *InvoicePaymentReceived) error {
	// TODO: implement
	if _, ok := s.payments[event.PaymentID]; !ok {
		data, ok := s.invoices[event.InvoiceID]
		if !ok {
			return errors.New("invoice not found")
		} else {
			s.payments[event.PaymentID] = InvoicePaymentReceived{}
			data.LastPaymentAt = event.PaidAt
			data.PaidAmount = decimal.Sum(data.PaidAmount, event.PaidAmount)
			data.FullyPaid = event.FullyPaid
			s.invoices[event.InvoiceID] = data
		}
	}

	return nil
}

func (s *InvoiceReadModelStorage) OnInvoiceVoided(ctx context.Context, event *InvoiceVoided) error {
	// TODO: implement
	data, ok := s.invoices[event.InvoiceID]
	if !ok {
		return errors.New("invoice not found")
	} else {
		data.VoidedAt = event.VoidedAt
		data.Voided = true
		s.invoices[event.InvoiceID] = data
	}

	return nil
}

func NewRouter(storage *InvoiceReadModelStorage,
	eventProcessorConfig cqrs.EventProcessorConfig,
	watermillLogger watermill.LoggerAdapter,
) (*message.Router, error) {
	router := message.NewDefaultRouter(watermillLogger)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessorConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create command processor: %w", err)
	}

	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"OnInvoiceIssued",
			storage.OnInvoiceIssued,
		),
		cqrs.NewEventHandler(
			"OnInvoicePaymentReceived",
			storage.OnInvoicePaymentReceived,
		),
		cqrs.NewEventHandler(
			"OnInvoiceVoided",
			storage.OnInvoiceVoided,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not add event handlers: %w", err)
	}

	return router, nil
}
