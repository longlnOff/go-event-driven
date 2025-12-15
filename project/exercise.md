# Project: Testing Ticket Cancellation

Now that we covered the happy path scenario, let's consider the ticket cancellation.

## Exercise

Exercise path: ./project

Based on the previous steps, **write tests for ticket cancellation.**

1. Call the API again with ticket's `status` set to `canceled`.
2. Assert that the ticket has been added to the `tickets-to-refund` sheet.

{{tip}}

When adding tests to existing functionality, it's a good practice to use the *test sabotage* technique.

Write the test, make it pass, and then break the code to see if the test fails.
This approach has saved us many times from having tests that weren't testing anything.

{{endtip}}
