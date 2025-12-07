# Project: Retries

The worker is now able to process messages in the background.
What if any errors happen while processing a message?

If we simply move to the next message, we **lose the current one.**
It's impossible to recover it later, so we end up with inconsistency in the system.

The common solution here is moving the message back to the queue and processing it again.

## Exercise

Exercise path: ./project

In this exercise, the `receipts` service and the spreadsheets API sometimes return `500 Internal Server Error`.
This is expected. The worker needs to handle that and retry the request.

**To make the worker retry failed messages, simply republish them using `Send` if an error happens.**
The message will be added to the queue and processed later.
This way, we don't lose messages while one of the APIs is down.

We can still lose the message if our service goes down while processing it.
Don't worry, we will improve this in the next module.

```go
if err != nil {
	w.Send(msg)
}
```
