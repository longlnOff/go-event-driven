# Project: Health checks

You can now control how the service shuts down. Another good practice is being able to tell when it's up and ready to serve requests.

The simple, common way to do this is to expose an HTTP endpoint like `/health` that returns a 200 status code once the service is ready.

Here's an example using Echo:

```go
e.GET("/health", func(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
})
```

The Watermill Router exposes a `Running` method. It returns a channel that gets closed once the Router is ready.
You can use it like this:

```go
<-router.Running()
```

## Exercise

Exercise path: ./project

**Implement a health check endpoint in your project.**

1. Expose an HTTP `GET /health` endpoint that returns a `200` status code and an `ok` message.

2. Extend your service code, so it waits for the Router to be ready before starting the HTTP server.
This way, the service is marked as healthy only when the Router is ready to process messages.

```go
g.Go(func() error {
	<-router.Running()
	
	err := e.Start(":8080")
	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}
	
	return nil
})
```
