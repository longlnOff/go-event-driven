package entities

import "github.com/google/uuid"

type BookTaxi struct {
	CustomerEmail      string `json:"customer_email"`
	CustomerName       string `json:"customer_name"`
	NumberOfPassengers int    `json:"number_of_passengers"`
	ReferenceID        string `json:"reference_id"`
	IdempotencyKey     string `json:"idempotency_key"`
}

type BookTaxiRequest struct {
	CustomerEmail      string
	NumberOfPassengers int
	PassengerName      string
	ReferenceId        string
	IdempotencyKey     string
}

type BookTaxiResponse struct {
	TaxiBookingId uuid.UUID `json:"taxi_booking_id"`
}
