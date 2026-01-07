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

type DataLakeEvent struct {
	EventID      string    `db:"event_id"`
	PublishedAt  time.Time `db:"published_at"`
	EventName    string    `db:"event_name"`
	EventPayload []byte    `db:"event_payload"`
}

type VipBundleInitialized_v1 struct {
	Header MessageHeader `json:"header"`

	VipBundleID VipBundleID `json:"vip_bundle_id"`
}

func (i VipBundleInitialized_v1) IsInternal() bool {
	return false
}

type BookShowTickets struct {
	BookingID uuid.UUID `json:"booking_id"`

	CustomerEmail   string    `json:"customer_email"`
	NumberOfTickets int       `json:"number_of_tickets"`
	ShowId          uuid.UUID `json:"show_id"`
}

func (i BookShowTickets) IsInternal() bool {
	return false
}

type BookFlight struct {
	CustomerEmail  string    `json:"customer_email"`
	FlightID       uuid.UUID `json:"to_flight_id"`
	Passengers     []string  `json:"passengers"`
	ReferenceID    string    `json:"reference_id"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func (i BookFlight) IsInternal() bool {
	return false
}

type FlightBooked_v1 struct {
	Header MessageHeader `json:"header"`

	FlightID  uuid.UUID   `json:"flight_id"`
	TicketIDs []uuid.UUID `json:"flight_tickets_ids"`

	ReferenceID string `json:"reference_id"`
}

func (i FlightBooked_v1) IsInternal() bool {
	return false
}

type VipBundleFinalized_v1 struct {
	Header MessageHeader `json:"header"`

	VipBundleID VipBundleID `json:"vip_bundle_id"`
	Success     bool        `json:"success"`
}

func (i VipBundleFinalized_v1) IsInternal() bool {
	return false
}

type BookingFailed_v1 struct {
	Header MessageHeader `json:"header"`

	BookingID     uuid.UUID `json:"booking_id"`
	FailureReason string    `json:"failure_reason"`
}

func (i BookingFailed_v1) IsInternal() bool {
	return false
}

type FlightBookingFailed_v1 struct {
	Header MessageHeader `json:"header"`

	FlightID      uuid.UUID `json:"flight_id"`
	FailureReason string    `json:"failure_reason"`

	ReferenceID string `json:"reference_id"`
}

func (i FlightBookingFailed_v1) IsInternal() bool {
	return false
}
