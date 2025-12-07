# Synchronous vs. Asynchronous

Most systems today use synchronous communication.
Whether it's HTTP with REST API or gRPC, the pattern is the same:
The client sends a request, and the server processes it and replies with a response.

Event-driven patterns take a different approach: asynchronous communication.

The synchronous request-reply communication is simple.
But there's also a drawback:
Your application needs to wait for the other server to complete the request.
This can take a long time, can be interrupted, or may fail for many reasons.
You could skip waiting for the response, but then you never know if the request was successful.

In most synchronous APIs, waiting for the request is the only way to know the result.
Problems start when you make multiple synchronous calls within one action.

Consider a classic web application.
When a user signs up, we want to save their account in the database, add them to the newsletter, and send them a welcome email.

```go
func SignUp(u User) error {
	if err := CreateUserAccount(u); err != nil {
		return err
	}
	
	if err := AddToNewsletter(u); err != nil {
		return err
	}
	
	if err := SendNotification(u); err != nil {
		return err
	}
	
	return nil
}
```

In this example, `AddToNewsletter` and `SendNotification` are HTTP requests over the network.
What happens when one of the calls fails for whatever reason?

You need to choose one of these options:

* Return an error to the user, roll back the database changes, and prevent them from signing up. Business-wise, this is a pretty bad outcome.
* Return a success to the user and ignore the errors. You end up with an inconsistency that needs to be fixed later — which means manual work.

We want to add users to the newsletter and send them a welcome email.
But these actions are not critical for the user to sign up and place an order.
**We don't need them to happen immediately, but we want them to happen eventually.**

This is where asynchronous patterns can help.

## Exercise

Exercise path: ./02-async/01-goroutines/main.go

Let's start with something naive that gets the job done.

In the exercise code, you will find the `SignUp` method, which is similar to the one above.

**The newsletter and notifications APIs are unstable and sometimes go down for unknown reasons.**
We don't want this to block users from signing up.
But we want users to be added to the newsletter and have a notification sent as soon as the APIs are back online.

{{tip}}

If you see a `network error`, it's not a problem with our platform — it's a part of the exercise!
Your task is to make the code resilient to such errors.

{{endtip}}

A trivial way to make a request asynchronous is running it in a goroutine.

```go
go func() {
	if err := AddToNewsletter(u); err != nil {
		log.Printf("failed to add user to the newsletter: %v", err)
	}
}()
```

We can add simple retries to make sure it eventually succeeds.

```go
go func() {
	for {
		if err := AddToNewsletter(u); err != nil {
			log.Printf("failed to add user to the newsletter: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}()
```

This is a naive approach: A simple restart of the service is enough to lose all the retries in progress.
But it's a good start to illustrate the idea.

**Let's start with a similar naive mechanism as in the above snippets and make `AddToNewsletter` and `SendNotification` async.**
Use retries for both the newsletter and notification APIs.

Keep in mind: **The `CreateUserAccount` method should stay synchronous.**
It's the only critical part of the sign-up process that must succeed before the user can continue.

Remember to add short sleeps between the retries!
