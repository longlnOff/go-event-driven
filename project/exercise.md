# Add ticket limits

The final feature we need to implement is limits on how many tickets we can sell.
We have this information in the `shows` table, so now we just need to enforce it.

{{background}}

For more popular shows, our API may receive up to 20 concurrent requests to book tickets for the same show.
We'll make sure that our system can handle this without overbooking.

{{endbackground}}

## Exercise

Exercise path: ./project

1. Enforce the limit of available tickets in the `POST /book-tickets` endpoint.

The endpoint should return `http.StatusBadRequest` if there are not enough tickets available.

It's fully up to you how you implement this logic.
The simplest approach may be to do it inside the repository method used to store the booking.
**You can get the number of available tickets, sum the already booked tickets, and check if the number of tickets to book is less than or equal to what's available.**

{{tip}}

We are using Echo in our example code. You can learn more about returning errors in the [Echo documentation](https://echo.labstack.com/docs/error-handling).

{{endtip}}

Make sure that you do this within the same transaction when storing the booking in the database.

2. You **must** use the `sql.LevelSerializable` isolation level to make sure no overbooking happens.
We will be aggregating tickets from multiple rows, so the repeatable read isolation level will be insufficient.

{{tip}}

You can read more about PostgreSQL serializable transactions [in the official documentation](https://www.postgresql.org/docs/13/transaction-iso.html#XACT-SERIALIZABLE)
and [this article](https://mkdev.me/posts/transaction-isolation-levels-with-postgresql-as-an-example).

{{endtip}}

We won't check your component tests, but we recommend implementing them.
This is a critical functionality, so it's good to have tests that ensure it works as expected.

You can check the {{exerciseLink "component testing" "08-component-tests" "03-project-running-service-in-tests"}} exercise for a reminder on how to run the tests locally.

{{hints}}

{{hint 1}}

This is how an example query for summing up the already booked tickets may look:

```sql
SELECT
	COALESCE(SUM(number_of_tickets), 0) AS already_booked_seats
FROM
	bookings
WHERE
	show_id = $1
```

{{endhint}}

{{hint 2}}

This is how an example extended implementation of the `AddBooking` method may look:

```go
func (b BookingsRepository) AddBooking(ctx context.Context, booking entities.Booking) (err error) {
	return updateInTx(
		ctx,
		b.db,
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {
			availableSeats := 0
			err = tx.GetContext(ctx, &availableSeats, `
				SELECT
					number_of_tickets AS available_seats
				FROM
					shows
				WHERE
					show_id = $1
			`, booking.ShowID)
			if err != nil {
				return fmt.Errorf("could not get available seats: %w", err)
			}

			alreadyBookedSeats := 0
			err = tx.GetContext(ctx, &alreadyBookedSeats, `
				SELECT
					COALESCE(SUM(number_of_tickets), 0) AS already_booked_seats
				FROM
					bookings
				WHERE
					show_id = $1
			`, booking.ShowID)
			if err != nil {
				return fmt.Errorf("could not get already booked seats: %w", err)
			}

			if availableSeats-alreadyBookedSeats < booking.NumberOfTickets {
				// this is usually a bad idea, learn more here: https://threedots.tech/post/introducing-clean-architecture/
				// we'll improve it later
				return echo.NewHTTPError(http.StatusBadRequest, "not enough seats available")
			}

			_, err = tx.NamedExecContext(ctx, `
				INSERT INTO 
					bookings (booking_id, show_id, number_of_tickets, customer_email) 
				VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email)
		`, booking)

		// ...	
```

{{endhint}}

{{endhints}}