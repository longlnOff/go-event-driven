# Rename topics

Currently, topic names are equal to command/event names.

Even if the chance of conflict is low, it would be good to separate them for clarity.
This will make it explicit whether the topic is for commands or events .

## Exercise

Exercise path: ./project

1. Change command and event topic names to the formats `events.{event_name}` and `commands.{command_name}`.
2. Change the consumer group names as well for consistency.

We will verify your solution by checking if you emit `TicketBookingConfirmed` on `events.TicketBookingConfirmed`
and `RefundTicket` on `commands.RefundTicket`.

{{tip}}

If you make such a change in production, you must be sure that there are no leftover messages on the old topics.

If you want to ensure that this doesn't happen, you can make the change in three steps:

1. Add handlers for new topics while keeping topics for old names. Deploy to production.
2. Change the command and event bus to publish to new topics. Deploy to production.
3. Make sure the old topics aren't used anymore and no messages are left on them.
You can do this by checking the message broker tools or checking logs/metrics.
4. Remove the handlers for old topics.

{{endtip}}
