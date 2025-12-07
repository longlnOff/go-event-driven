# Project: Introduce the Pub/Sub

It's time to return to our project and introduce a real Pub/Sub system.

We'll start with Redis, since it fits our requirements: it's lightweight and easy to configure.

{{tip}}

Keep in mind that Redis might or might not be a good choice for your real-world project.
Consider the tradeoffs and your needs before you commit to one of the Pub/Subs.

{{endtip}}

### Running the project locally _(optional)_

**You can complete the training without Docker installed, simply by using our CLI.**
However, if you want to **run the project locally**, you can do that as well.
This exercise includes a `docker-compose.yml` file to help you with that.

{{tip}}

If you don't have Docker installed on your machine, you can download it [here](https://www.docker.com/products/docker-desktop/).

{{endtip}}

First, run `docker compose up`. In another terminal, start your solution.
You will need to provide two environment variables:

```bash
# Mac or Linux
REDIS_ADDR=localhost:6379 GATEWAY_ADDR=http://localhost:8888 go run .

# Windows PowerShell
$env:REDIS_ADDR="localhost:6379"; $env:GATEWAY_ADDR="http://localhost:8888"; go run .
```

{{tip}}

When you run exercises with the `tdl` CLI, `REDIS_ADDR` and all required environment variables are set automatically.

{{endtip}}

The API should be up and running on `localhost:8080`.
You can now send a test API request:

```bash
curl -v -X POST http://localhost:8080/tickets-confirmation \
-H "Content-Type: application/json" \
-d '{"tickets": ["ticket-1", "ticket-2"]}'
```

## Exercise

Exercise path: ./project

Now it's time to apply the knowledge you've gained in the previous modules.
**Introduce the Redis Stream Pub/Sub to the project.**

Replace the worker implementation with Watermill publishers and subscribers.

Here are some tips to get you started:

* Create a `message` package in the `project` directory for publishers, subscribers, and the code handling messages.
* Inject the publisher into the HTTP handler in place of the worker.
* In the HTTP handler, instead of sending worker messages, publish Watermill messages on two topics: `issue-receipt` and `append-to-tracker`.
* Make TicketID the message payload (simply cast the string to `[]byte`).
* Create **two subscribers**, one for each topic. Each should use a unique consumer group.
* Iterate over incoming messages and execute the logic. Move the logic from the worker's `Run()` method.
* Watermill's `Message` exposes the context via the `Context()` method. Replace `context.Background()` with it.
* You can get rid of the rest of the worker code.

{{tip}}

You will continue working on this project throughout the training.
**Since the project will gain many features by the end, it's not worth cutting corners from the start.**
This approach will help you progress through the training more smoothly.

{{endtip}}

**Remember to run each iteration over messages in a separate goroutine.** Otherwise, you'll block the main function.

Don't forget about error handling! You should replace re-publishing logic with the `Ack()` and `Nack()` methods on the message.

```go
go func() {
	messages, err := sub.Subscribe(context.Background(), "topic")
	if err != nil {
		panic(err)
	}
	
	for msg := range messages {
		err := Action()
		if err != nil {
			msg.Nack()
		} else {
			msg.Ack()
		}
	}
}()
```

You may need to `go get` the dependencies:

```bash
go get github.com/ThreeDotsLabs/watermill
go get github.com/ThreeDotsLabs/watermill-redisstream
go get github.com/redis/go-redis/v9
```

{{tip}}

We recommend submitting your solution early and often to get quick feedback from our tests.

If you get stuck, you can always ask the Mentor or your peers on the Discord channel.

You can also see an example solution after submitting an invalid solution multiple times.

{{endtip}}
