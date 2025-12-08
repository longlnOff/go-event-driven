# Project: Accepting more details

{{background}}

Our integrations work correctly, but adding just the ticket IDs to the spreadsheet isn't that helpful.
The operations team still has to manually look up the ticket details.
We can help them out by adding more details in the spreadsheet.

{{endbackground}}

First, we're going to extend the webhook endpoint to accept more than ticket IDs.
We're dropping support for `POST /tickets-confirmation`.
Instead, we're going to expose `POST /tickets-status`.

Previously, the endpoint accepted a list of ticket IDs; now it's going to accept a list of tickets.
Each ticket has a ticket ID, a status, a customer email, and a price.

For now, we only support `confirmed` tickets. We'll add other statuses later.

Here's an example incoming HTTP request:

```json
{
  "tickets": [
    {
      "ticket_id": "ticket-1",
      "status": "confirmed",
      "customer_email": "user@example.com",
      "price": {
        "amount": "50.00",
        "currency": "EUR"
      }
    }
  ]
}
```

Go structs can look like this:

```go
type TicketsStatusRequest struct {
	Tickets []TicketStatusRequest `json:"tickets"`
}

type TicketStatusRequest struct {
	TicketID      string `json:"ticket_id"`
	Status        string `json:"status"`
	Price         Money  `json:"price"`
	CustomerEmail string `json:"customer_email"`
}
```

We can also add a `Money` type in the `entities` package:
It will be useful in many places later.

```go
package entities

type Money struct {
    Amount   string `json:"amount"`
    Currency string `json:"currency"`
}
```

{{tip}}

Note that we keep the money's amount as a `string` instead of a `float64`.
This is intentional.
Don't use `float64` for money, as you may lose precision.

In this training, we won't perform any calculations on the price, so using a `string` is fine.
If you need to perform calculations, use a library like [github.com/shopspring/decimal](https://github.com/shopspring/decimal).

{{endtip}}

## Exercise

Exercise path: ./project

**Replace the `POST /tickets-confirmation` endpoint with `POST /tickets-status`.**

The handler should work the same.
Publish messages with the ticket ID as the payload.
We'll add more details in the next exercise.
