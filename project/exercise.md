# Removing canceled tickets

Did you remember that it's possible to cancel tickets?
We should reflect this in our API as well.

{{background}}

The team that requested the API said that they don't need information about ticket status —
 they just need to know the tickets that were not canceled.

{{endbackground}}

## Exercise

Exercise path: ./project

Implement a handler that listens for the `TicketBookingCanceled` event and removes canceled tickets from the database.

We don't need to use soft delete here—we can simply remove the ticket from the database.
