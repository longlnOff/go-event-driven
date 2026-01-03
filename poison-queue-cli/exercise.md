# Poison Queue CLI: Requeue

Sometimes the issue is not with the message but with the handler code.
In this scenario, we can release a new version and then republish (requeue) the poisoned message to the original topic. 

## Exercise

Exercise path: ./poison-queue-cli

Similar to the `Remove` method, iterate over the messages and look for the given message ID.

Acknowledge it to remove it from the poison queue, but first publish it on the original topic.

You can get the original topic from the metadata as described in {{exerciseLink "the previous exercise" "20-fault-tolerance" "05-poison-queue-cli-preview"}}.

If the message is not found, return an error.

{{tip}}

Never ack the messages before they're successfully re-published — you could lose them!

If you need to choose, it's better to deal with duplicates of the same message rather than losing the message.

You can read more in {{exerciseLink "the at-least-once delivery module" "10-at-least-once-delivery" "05-idempotent-event-handlers"}}
on how to handle duplicates in your event handlers.

{{endtip}}
