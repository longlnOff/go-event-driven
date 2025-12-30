package entities

import (
	"time"

	"github.com/google/uuid"
)

type OpsBooking struct {
	BookingID uuid.UUID `json:"booking_id"` // from BookingMade event
	BookedAt  time.Time `json:"booked_at"`  // from BookingMade event

	Tickets map[string]OpsTicket `json:"tickets"` // Tickets added/updated by TicketBookingConfirmed, TicketRefunded, TicketPrinted, TicketReceiptIssued

	LastUpdate time.Time `json:"last_update"` // updated when read model is updated
}

type OpsTicket struct {
	PriceAmount   string `json:"price_amount"`   // from TicketBookingConfirmed event
	PriceCurrency string `json:"price_currency"` // from TicketBookingConfirmed event
	CustomerEmail string `json:"customer_email"` // from TicketBookingConfirmed event

	ConfirmedAt time.Time `json:"confirmed_at"` // from TicketBookingConfirmed event
	RefundedAt  time.Time `json:"refunded_at"`  // from TicketRefunded event

	PrintedAt       time.Time `json:"printed_at"`        // from TicketPrinted event
	PrintedFileName string    `json:"printed_file_name"` // from TicketPrinted event

	ReceiptIssuedAt time.Time `json:"receipt_issued_at"` // from TicketReceiptIssued event
	ReceiptNumber   string    `json:"receipt_number"`    // from TicketReceiptIssued event
}
