# Handling out-of-order `TicketBookingConfirmed` and `TicketBookingCanceled` events

We can't guess how you implemented removing tickets from the `tickets` table.
However, there is a chance that your code doesn't assume that 
`TicketBookingConfirmed` and `TicketBookingCanceled` events can arrive out of order.

In our previous example solution, we were using a query like this one:

```sql
DELETE FROM tickets WHERE ticket_id = $1
```

It will execute successfully even if there is no row for the ticket in the table yet.
**However, some tickets may be not removed if the `TicketBookingCanceled` event is 
processed before `TicketBookingConfirmed` because there is nothing to remove.**

In theory, we could check if there is a row for the ticket in the table before removing it, and return an error if there isn't.
However, there is one downside: Our repository will be no longer be {{exerciseLink "idempotent" "10-at-least-once-delivery" "05-idempotent-event-handlers"}}.
**If we receive the `TicketBookingCanceled` event again, it will fail because there will be no row for the ticket in the table.**

One of the solutions here is to use soft delete.
We need to add a `deleted_at` column to the `tickets` table and exclude those rows from the result when querying for tickets.

{{tip}}

A repository pattern is pretty useful in such cases — you have one central place where you can change your querying logic.
When you are not querying your data in multiple places in your codebase,
you don't need to be afraid that you will forget to update your logic somewhere.

Bonus points for writing {{exerciseLink "integration tests" "10-at-least-once-delivery" "07-project-testing-idempotency"}} for that logic!

{{endtip}}

When `TicketBookingCanceled` arrives before `TicketBookingConfirmed` (in other words, when no ticket was stored in `tickets` yet), 
we will return an error. `TicketBookingCanceled` will be redelivered after a while, when the ticket should already exist.
Upon redelivery of `TicketBookingCanceled`, we will just ignore the update.
In other words, our repository will be idempotent and resilient to out-of-order events.

## Exercise

Exercise path: ./project

1. Add a `deleted_at` column to the `tickets` table.
2. Set `deleted_at` to the current timestamp when the `TicketBookingCanceled` event is processed.
3. Update your querying logic, so tickets with non-null `deleted_at` aren't returned.

4. We will simulate sending `POST /tickets-status` with an out-of-order status update where you receive cancellation information for a ticket not yet booked.
**Check if the booking exists in `tickets` and return an error in your event handler when `deleted_at` is not set. This prevents losing the event.**
This will nack the message for redelivery after `TicketBookingConfirmed` is processed.

{{tip}}

Ensure {{exerciseLink "your retry middleware" "07-errors" "02-project-temporary-errors"}} doesn't redeliver messages too slowly.
Messages redelivered after too long may exceed test timeout.

In the end, canceled tickets should not be returned from your `GET /tickets` endpoint.

{{endtip}}


{{hints}}

{{hint 1}}

It's example code of how soft delete with ensuring that booking exists can be implemented:

```go
func (t TicketsRepository) Remove(ctx context.Context, ticketID string) error {
	res, err := t.db.ExecContext(
		ctx,
		`UPDATE tickets SET deleted_at = now() WHERE ticket_id = $1`,
		ticketID,
	)
	if err != nil {
		return fmt.Errorf("could not remove ticket: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("ticket with id %s not found", ticketID)
	}

	return nil
}
```

{{endhint}}

{{endhints}}