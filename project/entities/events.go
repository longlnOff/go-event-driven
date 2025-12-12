package entities

import (
	"time"

	"github.com/google/uuid"
)

type MessageHeader struct {
	ID          string    `json:"id"`
	PublishedAt time.Time `json:"published_at"`
}

func NewMessageHeader() MessageHeader {
	return MessageHeader{
		ID:          uuid.NewString(),
		PublishedAt: time.Now().UTC(),
	}
}

type TicketBookingConfirmed struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}
