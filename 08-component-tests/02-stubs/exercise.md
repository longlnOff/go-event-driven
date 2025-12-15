# Stubs

You need to be able to do all three of the following to execute component tests:

- Run the service as a function
- Stub out external dependencies
- Use infrastructure that runs in Docker

We will cover all of these in this module.

Let's start with writing stubs for our external dependencies.
Remember that we don't consider Redis as an *external dependency*.
We can run it in Docker and use the same code to work with it.

**The key is hiding dependencies behind interfaces.**

Luckily, we already do it in the project.
For example, with `ReceiptsService`:

```go
type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) error
}
```

The implementation we currently use calls a remote HTTP API.
We can implement the same interface to create a stub implementation.

Note that we don't return or accept any types from the common library's API clients except `entities.IssueReceiptRequest`.
It's a good practice to not propagate any external dependencies to the rest of the application.
Any breaking changes to our API client would force us to make changes in multiple places.
If we isolate it, we only need to change the adapter.

It's also a good idea to not expose all the data that we get from external APIs, just the data that we need.
This decreases coupling and makes our code simpler.

### Stubs vs Mocks

There are many libraries that provide very powerful mocks.

Usually, it works like this:

```go
mock := NewMock()
mock.ExpectMethod().Save().WithArguments("user").ToReturn(nil)
```

**We do not recommend using this approach in tests.**
Such mocks are often fragile and hard to debug.
Each time your logic changes, you will need to update the mocks.

What's worse, they give you false confidence that your tests do something meaningful.
In reality, you test method calls, not the behavior.

**Instead, we suggest writing stubs.**
It's usually quick and gives you a lot of flexibility.

A stub is simply a struct that implements the given interface.
Instead of calling an external API, you fake the behavior inside it.
It can be as simple as returning a hardcoded value or storing the passed arguments for later assertions.

### Implementation

This is how the stub implementation might look:

```go
type Input struct {
	TicketID string
}

type Result struct {
	Number string
	DoneAt time.Time
}

type Stub struct {
	// let's make it thread-safe
	lock sync.Mutex

	Inputs []Input
}

func (s *Stub) DoStuff(ctx context.Context, input Input) (Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Inputs = append(s.Inputs, input)

	return Result{
        Number: "fake-number",
		DoneAt: time.Now(),
	}, nil
}
```

It's possible to write such a stub in a minute.

We have a lot of flexibility about how the stub should work.
We can also model any external logic.

The benefit compared to a generated mock is that we write the stub logic only once, not for each test.
We also don't need to define which function calls we expect.

In tests, we can assert what inputs have been passed by accessing the `Inputs` field.

This approach assumes that external services behave in the way that you implemented in the mock.
It's not possible to ensure that this is always true, but it's also not a promise of tests in general.
Tests should give you more confidence that your code works as expected, but they are not guarantees.

## Exercise

Exercise path: ./08-component-tests/02-stubs/main.go

Write a stub for the provided `ReceiptsService` interface.

It should be a struct called `ReceiptsServiceStub`.

It should have a public field `IssuedReceipts []IssueReceiptRequest` that stores all the requests that were passed to the `IssueReceipt` method.
The stub's `IssueReceipt` should simply return `nil`.

`ReceiptsServiceStub` should be thread-safe.
It will be used later in component tests that are executed in parallel.
To ensure this, use a mutex like this:

```go
func (s *Stub) SomeMethod() {
	s.lock.Lock()
	defer s.lock.Unlock()
	
	// ...
```

{{tip}}

Don't forget to use a pointer receiver for the stub's `IssueReceipt` method.
Otherwise, it won't be able to modify any struct fields.

If your code doesn't work and there's no obvious reason, check this first.

```go
func (s *ReceiptsServiceStub) IssueReceipt(
	ctx context.Context,
	request IssueReceiptRequest,
) error {
```

{{endtip}}
