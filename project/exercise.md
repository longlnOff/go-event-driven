# Adding a Process Manager to the Project

The core of the process manager is ready. It's time to connect it to the rest of our project.

In this exercise, we will use many things we learned earlier:

- {{exerciseLink "outbox" "11-outbox" "05-publishing-events-in-transaction"}}: We need to publish `VipBundleInitialized_v1` in a transaction.
- {{exerciseLink "commands" "12-cqrs-commands" "01-command-vs-event"}}: For booking show tickets, and then later, flight.
- {{exerciseLink "tracing" "19-tracing" "01-tracing"}}: Will help us debug and understand what is going on.
- {{exerciseLink "at-least-once-delivery and deduplication" "10-at-least-once-delivery" "05-idempotent-event-handlers"}}: All our command and event handlers within the process manager should be idempotent. We don't want to book flight tickets twice.

## Exercise

Exercise path: ./project

**We're back to the "tickets" project. From now on, work in your `project` directory.**

**In this exercise, we'll verify the happy path of booking show tickets.**
We won't verify booking flights or handling rollbacks yet. We'll do it in the next exercise.

For now, we just want to initialize the VIP Bundle and book show tickets.

**You should now use the process manager you have implemented in the previous exercise in your project.**

1. Add an HTTP endpoint `POST /book-vip-bundle` that creates the VipBundle and stores it in the database.

Use this request body:

```go
type vipBundleRequest struct {
	CustomerEmail   string    `json:"customer_email"`
	InboundFlightID uuid.UUID `json:"inbound_flight_id"`
	NumberOfTickets int       `json:"number_of_tickets"`
	Passengers      []string  `json:"passengers"`
	ShowID          uuid.UUID `json:"show_id"`
}
```

And this response:

```go
type vipBundleResponse struct {
	BookingID   uuid.UUID   `json:"booking_id"`
	VipBundleID VipBundleID `json:"vip_bundle_id"`
}
```

The endpoint should emit `VipBundleInitialized_v1`.
It should be done within a transaction. Use {{exerciseLink "the outbox pattern" "11-outbox" "05-publishing-events-in-transaction"}} to publish the event.

You can take inspiration from previous exercises' tests on how to initialize the VIP Bundle.

The endpoint should return `StatusCreated (201)` if the process manager starts successfully.

Generate the VIP bundle ID and booking ID in the HTTP handler.

{{tip}}

**It's a common approach to generate IDs of entities when sending commands.**

For example:

```go
vb := entities.VipBundle{
    VipBundleID:     entities.VipBundleID{uuid.New()},
    BookingID:       uuid.New(),
    // ...
}
```

**This helps track the entity in the system later.**
**We can return the VIP bundle and Booking ID from the API response. This works even if the entity is created asynchronously.**

```go
return c.JSON(http.StatusCreated, vipBundleResponse{
     BookingID:   vb.BookingID,
     VipBundleID: vb.VipBundleID,
 })
```

As a downside, it doesn't work well with auto-increment IDs.
But in practice, UUID v7 solves this problem.
UUID is storage efficient when stored as binary or UUID type.
It is also time-sortable, so it doesn't cause slow inserts.

{{endtip}}

2. Implement a repository for `VipBundle`. It should store the bundle in the database and emit `VipBundleInitialized_v1`.

You can keep the `VipBundleRepository` interface from the previous exercises or redesign it as you see fit.

The interface is inspired by our [article on the repository pattern](https://threedots.tech/post/repository-pattern-in-go/).

You can see example implementation in the _Hints_ section below.

3. Implement the `BookShowTickets` command handler. It should call the existing logic used by the `POST /book-tickets` endpoint.

While the endpoint generates the booking ID, the command handler should use the `BookingID` field from the command.

Beware: Previously, tickets were booked just by the HTTP endpoint. We didn't need to deduplicate them.

Now a command will also use this logic. Commands may be {{exerciseLink "redelivered (at-least-once delivery, do you remember?)" "10-at-least-once-delivery" "05-idempotent-event-handlers"}}.

You can use this helper to detect when the booking has already been done:

```go
const (
	postgresUniqueValueViolationErrorCode = "23505"
)

func isErrorUniqueViolation(err error) bool {
	var psqlErr *pq.Error
	return errors.As(err, &psqlErr) && psqlErr.Code == postgresUniqueValueViolationErrorCode
}
```

Check if the error is a unique violation error when inserting a booking.
If it is, return a special error. This error indicates the booking already exists.

```go
var (
	ErrBookingAlreadyExists = errors.New("booking already exists")
)

// ...

if isErrorUniqueViolation(err) {
	// now AddBooking is called via Pub/Sub, we need to take into account at-least-once delivery
	return ErrBookingAlreadyExists
}

// ...


// ...

if errors.Is(err, ErrBookingAlreadyExists) {

// ...
```

**IMPORTANT: to make it work, you must try to insert the booking before checking available tickets.**
If you check available tickets first, a re-delivery may show insufficient tickets.
This happens even if the booking has already been completed in the previous delivery.

If this happens in your command handler, return `nil` error to acknowledge the command.
The HTTP handler should work as-is and return an error.

4. Create a new `VipBundleProcessManager` instance. Add it to the message router.

Copy and adapt entities and events from the previous exercise.

Add these event handlers to the event processor:

- `OnVipBundleInitialized`
- `OnBookingMade`
- `OnTicketBookingConfirmed`
- `OnBookingFailed`
- `OnFlightBooked`
- `OnFlightBookingFailed`

**Remember, in this exercise, the tests check the happy path only. We won't verify booking flights yet. We'll do it in the next exercise.**

{{tip}}

This exercise can benefit from using [Clean Architecture](https://threedots.tech/post/introducing-clean-architecture/).
You can use the same function for both HTTP handler and command handler to book show tickets.

{{endtip}}

Don't worry if this doesn't work the first time.
If something breaks, check the logs. Try to understand what is happening.

### Alerting

We won't check it in this exercise, but in a production system, you may want to add alerting for stuck process managers.

You can periodically query process manager instances in the database and check which are not completed within the provided threshold.
It should also be possible to check stuck process managers out of events in the data lake.

### Testing

We won't check this, but we recommend adding {{exerciseLink "component tests" "08-component-tests" "01-component-tests"}} for this endpoint and for the process manager.

{{hints}}

{{hint 1}}

Do you need inspiration for how to implement `VipBundleRepository`?

You can use this schema:

```sql
CREATE TABLE IF NOT EXISTS vip_bundles (
	vip_bundle_id UUID PRIMARY KEY,
	booking_id UUID NOT NULL UNIQUE,
	payload JSONB NOT NULL
); 
```

And use this code:

```go
type VipBundleRepository struct {
	db *sqlx.DB
}

func NewVipBundleRepository(db *sqlx.DB) *VipBundleRepository {
	if db == nil {
		panic("db must be set")
	}

	return &VipBundleRepository{db: db}
}

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (v VipBundleRepository) Add(ctx context.Context, vipBundle entities.VipBundle) error {
	payload, err := json.Marshal(vipBundle)
	if err != nil {
		return fmt.Errorf("could not marshal vip bundle: %w", err)
	}

	return updateInTx(
		ctx,
		v.db,
		sql.LevelRepeatableRead,
		func(ctx context.Context, tx *sqlx.Tx) error {
			_, err = v.db.ExecContext(ctx, `
				INSERT INTO vip_bundles (vip_bundle_id, booking_id, payload)
				VALUES ($1, $2, $3)
			`, vipBundle.VipBundleID, vipBundle.BookingID, payload)

			if err != nil {
				return fmt.Errorf("could not insert vip bundle: %w", err)
			}

			outboxPublisher, err := outbox.NewPublisherForDb(ctx, tx)
			if err != nil {
				return fmt.Errorf("could not create event bus: %w", err)
			}

			err = event.NewBus(outboxPublisher).Publish(ctx, entities.VipBundleInitialized_v1{
				Header:      entities.NewMessageHeader(),
				VipBundleID: vipBundle.VipBundleID,
			})
			if err != nil {
				return fmt.Errorf("could not publish event: %w", err)
			}

			return nil
		},
	)
}

func (v VipBundleRepository) Get(ctx context.Context, vipBundleID entities.VipBundleID) (entities.VipBundle, error) {
	return v.vipBundleByID(ctx, vipBundleID, v.db)
}

func (v VipBundleRepository) vipBundleByID(ctx context.Context, vipBundleID entities.VipBundleID, db Executor) (entities.VipBundle, error) {
	var payload []byte
	err := v.db.QueryRowContext(ctx, `
		SELECT payload FROM vip_bundles WHERE vip_bundle_id = $1
	`, vipBundleID).Scan(&payload)

	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not get vip bundle: %w", err)
	}

	var vipBundle entities.VipBundle
	err = json.Unmarshal(payload, &vipBundle)
	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not unmarshal vip bundle: %w", err)
	}

	return vipBundle, nil
}

func (v VipBundleRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (entities.VipBundle, error) {
	return v.getByBookingID(ctx, bookingID, v.db)
}

func (v VipBundleRepository) getByBookingID(ctx context.Context, bookingID uuid.UUID, db Executor) (entities.VipBundle, error) {
	var payload []byte
	err := db.QueryRowContext(ctx, `
		SELECT payload FROM vip_bundles WHERE booking_id = $1
	`, bookingID).Scan(&payload)

	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not get vip bundle: %w", err)
	}

	var vipBundle entities.VipBundle
	err = json.Unmarshal(payload, &vipBundle)
	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not unmarshal vip bundle: %w", err)
	}

	return vipBundle, nil
}

func (v VipBundleRepository) UpdateByID(ctx context.Context, vipBundleID entities.VipBundleID, updateFn func(vipBundle entities.VipBundle) (entities.VipBundle, error)) (entities.VipBundle, error) {
	var vb entities.VipBundle

	err := updateInTx(ctx, v.db, sql.LevelSerializable, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		vb, err = v.vipBundleByID(ctx, vipBundleID, tx)
		if err != nil {
			return err
		}

		vb, err = updateFn(vb)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(vb)
		if err != nil {
			return fmt.Errorf("could not marshal vip bundle: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE vip_bundles SET payload = $1 WHERE vip_bundle_id = $2
		`, payload, vb.VipBundleID)

		if err != nil {
			return fmt.Errorf("could not update vip bundle: %w", err)
		}

		return nil
	})
	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not update vip bundle: %w", err)
	}

	return vb, nil
}

func (v VipBundleRepository) UpdateByBookingID(ctx context.Context, bookingID uuid.UUID, updateFn func(vipBundle entities.VipBundle) (entities.VipBundle, error)) (entities.VipBundle, error) {
	var vb entities.VipBundle

	err := updateInTx(ctx, v.db, sql.LevelSerializable, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		vb, err = v.getByBookingID(ctx, bookingID, tx)
		if err != nil {
			return err
		}

		vb, err = updateFn(vb)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(vb)
		if err != nil {
			return fmt.Errorf("could not marshal vip bundle: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE vip_bundles SET payload = $1 WHERE booking_id = $2
		`, payload, vb.BookingID)

		if err != nil {
			return fmt.Errorf("could not update vip bundle: %w", err)
		}

		return nil
	})
	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not update vip bundle: %w", err)
	}

	return vb, nil
}
```

{{endhint}}

{{hint 2}}

Don't forget about adding the command and event handlers to your event processor:

```go
// ...
cqrs.NewEventHandler(
   "vip_bundle_process_manager.OnVipBundleInitialized",
   vipBundleProcessManager.OnVipBundleInitialized,
),
cqrs.NewEventHandler(
   "vip_bundle_process_manager.OnBookingMade",
   vipBundleProcessManager.OnBookingMade,
),
cqrs.NewEventHandler(
   "vip_bundle_process_manager.OnTicketBookingConfirmed",
   vipBundleProcessManager.OnTicketBookingConfirmed,
),
cqrs.NewEventHandler(
   "vip_bundle_process_manager.OnBookingFailed",
   vipBundleProcessManager.OnBookingFailed,
),
cqrs.NewEventHandler(
   "vip_bundle_process_manager.OnFlightBooked",
   vipBundleProcessManager.OnFlightBooked,
),
cqrs.NewEventHandler(
   "vip_bundle_process_manager.OnFlightBookingFailed",
   vipBundleProcessManager.OnFlightBookingFailed,
),

// ...

cqrs.NewCommandHandler(
   "BookShowTickets",
   commandsHandler.BookShowTickets,
),

// ...
```

{{endhint}}

{{hint 3}}

Here's how an example HTTP handler for `POST /book-vip-bundle` endpoint looks like:

```go
package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"tickets/entities"
)

type vipBundleRequest struct {
	CustomerEmail   string    `json:"customer_email"`
	InboundFlightID uuid.UUID `json:"inbound_flight_id"`
	NumberOfTickets int       `json:"number_of_tickets"`
	Passengers      []string  `json:"passengers"`
	ShowID          uuid.UUID `json:"show_id"`
}

type vipBundleResponse struct {
	BookingID   uuid.UUID `json:"booking_id"`
	VipBundleID entities.VipBundleID `json:"vip_bundle_id"`
}

func (h Handler) PostVipBundle(c echo.Context) error {
	var request vipBundleRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	if request.NumberOfTickets < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "number of tickets must be greater than 0")
	}

	vb := entities.VipBundle{
		VipBundleID:     entities.VipBundleID{uuid.New()},
		BookingID:       uuid.New(),
		CustomerEmail:   request.CustomerEmail,
		NumberOfTickets: request.NumberOfTickets,
		ShowID:          request.ShowID,
		Passengers:      request.Passengers,
		InboundFlightID: request.InboundFlightID,
		IsFinalized:     false,
		Failed:          false,
	}

	if err := h.vipBundlesRepository.Add(c.Request().Context(), vb); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, vipBundleResponse{
		BookingID:   vb.BookingID,
		VipBundleID: vb.VipBundleID,
	})
}
```

{{endhint}}

{{endhints}}
