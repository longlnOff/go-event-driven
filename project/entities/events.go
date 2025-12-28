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

type TicketBookingConfirmed struct {
	Header        MessageHeader `json:"header"`
	BookingID     string        `json:"booking_id"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         Money         `json:"price"`
}

type TicketBookingCanceled struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         Money         `json:"price"`
}

type TicketPrinted struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

type BookingMade struct {
	Header MessageHeader `json:"header"`

	NumberOfTickets int    `json:"number_of_tickets"`
	BookingID       string `json:"booking_id"`
	CustomerEmail   string `json:"customer_email"`
	ShowID          string `json:"show_id"`
}

type TicketReceiptIssued struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	ReceiptNumber string `json:"receipt_number"`

	IssuedAt time.Time `json:"issued_at"`
}

type TicketRefunded struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}
