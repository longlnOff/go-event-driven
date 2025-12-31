package entities

import (
	"time"

	"github.com/google/uuid"
)

type MessageHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func NewMessageHeader() MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: uuid.NewString(),
	}
}

func NewMessageHeaderWithIdempotencyKey(idempotencyKey string) MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: idempotencyKey,
	}
}

type TicketBookingConfirmed_v1 struct {
	Header        MessageHeader `json:"header"`
	BookingID     string        `json:"booking_id"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         Money         `json:"price"`
}

func (i TicketBookingConfirmed_v1) IsInternal() bool {
	return false
}

type TicketBookingCanceled_v1 struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         Money         `json:"price"`
}

func (i TicketBookingCanceled_v1) IsInternal() bool {
	return false
}

type TicketPrinted_v1 struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

func (i TicketPrinted_v1) IsInternal() bool {
	return false
}

type BookingMade_v1 struct {
	Header MessageHeader `json:"header"`

	NumberOfTickets int    `json:"number_of_tickets"`
	BookingID       string `json:"booking_id"`
	CustomerEmail   string `json:"customer_email"`
	ShowID          string `json:"show_id"`
}

func (i BookingMade_v1) IsInternal() bool {
	return false
}

type TicketReceiptIssued_v1 struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	ReceiptNumber string `json:"receipt_number"`

	IssuedAt time.Time `json:"issued_at"`
}

func (i TicketReceiptIssued_v1) IsInternal() bool {
	return false
}

type TicketRefunded_v1 struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}

func (i TicketRefunded_v1) IsInternal() bool {
	return false
}

type InternalOpsReadModelUpdated struct {
	Header MessageHeader `json:"header"`

	BookingID uuid.UUID `json:"booking_id"`
}

func (i InternalOpsReadModelUpdated) IsInternal() bool {
	return true
}

type Event interface {
	IsInternal() bool
}

// We just need to unmarshal the event header; the rest is stored as is.
type ExternalEvent struct {
	Header MessageHeader `json:"header"`
}
