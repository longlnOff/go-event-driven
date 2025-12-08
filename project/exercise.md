# Project: The Router

## Exercise

Exercise path: ./project

**Rework your project to use the Router.**

1. Remove the raw `Subscribe` calls and iteration over the messages.
2. Replace them with Router's `AddConsumerHandler` methods.

You can name your handlers whatever you want but keep the same topic names.
Keep two subscribers, each with its own consumer group; each handler should use one of the subscribers.

**Remember: When using the Router, you shouldn't call `Ack()` or `Nack()` explicitly.**
Instead, return the proper error from the handler function.

3. **To run the Router, to call `Run` in a separate goroutine.**
Do it in the `service` package in the Service's `Run()` method.

```go
go func() {
	err := router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}()
```
