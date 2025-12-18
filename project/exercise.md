# Update component tests

Did you remember to update the component tests?
If not... it's time to do that now!

{{tip}}

Since the {{exerciseLink "component tests module" "08-component-tests" "01-component-tests"}} was optional,
you may have skipped it.

Would you like to go back to the component tests module and check it out?
You can use `tdl tr jump` to jump to the module and complete it now.

{{endtip}}


We won't check whether your test work.
It's up to you if you want to fix them.

You can find instructions on how to run component tests locally in the {{exerciseLink "Running the Service in Tests" "08-component-tests" "03-project-running-service-in-tests"}} exercise.

Since we have added a database support, now we need to also pass the database connection URL to run tests:

```bash
# Mac or Linux
REDIS_ADDR=localhost:6379 POSTGRES_URL=postgres://user:password@localhost:5432/db?sslmode=disable go test ./tests/ -v

# Windows PowerShell
$env:REDIS_ADDR="localhost:6379"; $env:POSTGRES_URL="postgres://user:password@localhost:5432/db?sslmode=disable"; go test ./tests/ -v
```

## Exercise

Exercise path: ./project

1. Implement stub of Files API and inject it into service.
2. Test idempotency of the `sendTicketsStatus` function by sending the same request multiple times with the same idempotency key.
3. Check that tickets were printed by calling Files API
4. Pass idempotency key calls of POST `/tickets-status`
5. Check if ticket was stored in the database


If you want, you can spend some time on checking the idempotency of some scenarios that are not possible to test at the repository level, 
such as issuing receipts. It's critical to make sure that receipts are issued only once for each ticket.
We don't want to mess with the financial team, do we?


{{hints}}

{{hint 1}}

This is how example implementation function that checks if ticket was stored in the repository looks like:

```go
func assertTicketStoredInRepository(t *testing.T, db *sqlx.DB, ticket ticketsHttp.TicketStatusRequest) {
	ticketsRepo := dbAdapters.NewTicketsRepository(db)

	assert.Eventually(
		t,
		func() bool {
			tickets, err := ticketsRepo.FindAll(context.Background())
			if err != nil {
				return false
			}

			for _, t := range tickets {
				if t.TicketID == ticket.TicketID {
					return true
				}
			}

			return false
		},
		10*time.Second,
		100*time.Millisecond,
	)
}
```

{{endhint}}

{{endhints}}
