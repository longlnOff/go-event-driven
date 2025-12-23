# Catching the Book Ticket request

So far, we operated on single tickets.
But customers can book multiple tickets at once, so we need to introduce a new concept in our system: **booking**.

A booking consists of:

- Booking ID
- Show ID
- Number of tickets
- Customer email

Next, we're going to implement the `POST /book-tickets` endpoint.
It mimics the Dead Nation API.

This endpoint should accept requests in this format:

```json
{
  "show_id": "0299a177-a68a-47fb-a9fb-7362a36efa69",
  "number_of_tickets": 3,
  "customer_email": "email@example.com"
}
```

You can use this Go structure as a starting point:

```go
type Booking struct {
    BookingID       string `json:"booking_id" db:"booking_id"`
    ShowID          string `json:"show_id" db:"show_id"`
    NumberOfTickets int    `json:"number_of_tickets" db:"number_of_tickets"`
    CustomerEmail   string `json:"customer_email" db:"customer_email"`
}
```

## Exercise

Exercise path: ./project

**Implement the `POST /book-tickets` endpoint.**

It should store bookings in the `bookings` table. 
(It's important to use this table name: We'll use it to check your solution.)

You can use a schema like this:

```sql
CREATE TABLE IF NOT EXISTS bookings (
    booking_id UUID PRIMARY KEY,
    show_id UUID NOT NULL,
    number_of_tickets INT NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    FOREIGN KEY (show_id) REFERENCES shows(show_id)
);
```

The booking ID is not sent in the request — you should generate it on your side.

The booking ID should be returned as the response from this endpoint with status code `201 Created`.

```json
{
    "booking_id": "bde0bd8d-88df-4872-a099-d4cf5eb7b491"
}
```

{{tip}}

Consider creating a separate database repository for the `Booking` entity.

{{endtip}}
