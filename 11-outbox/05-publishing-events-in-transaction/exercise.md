# Publishing Events Within a Transaction

Let's implement a simple outbox.
We'll use a bit of Watermill's help to achieve it.

{{tip}}

At first glance, this may sound like a simple task.
It's just inserting events and querying one by one and publishing them, right?
Yes, but the devil is in the details.

The biggest problem to handle is the situation where transactions are committed in a different order than they were started.
This leads to non-ordered message IDs, which makes consuming more difficult than sending messages one by one.
The problem is illustrated by [this test](https://github.com/ThreeDotsLabs/watermill-sql/blob/ea4d755fa8bb7a30cf252c0d5799dbebe2357731/pkg/sql/pubsub_test.go#L297).

{{endtip}}

For the sake of the training, we'll use the Watermill SQL Pub/Sub.
The SQL Pub/Sub has the same interface as other Pub/Subs. It uses a SQL database (PostgreSQL or MySQL) as its backend.

{{tip}}

The SQL Pub/Sub has one interesting advantage: It's possible to use it as a replacement for a "real" message broker.
In some cases, due to technical (or political) limitations in your company, it's **not easy to
deploy a message broker in your infrastructure.**

**In such cases, the SQL Pub/Sub can be used as a good replacement for a message broker.**
This can be a nice ice-breaker for introducing event-driven architecture in your company.
After you show that it's valuable, it can simplify the discussion about having a "real" message broker.

However, it may also be fine to just stay with the SQL Pub/Sub.
We know one famous company that uses just an SQL Pub/Sub in production, and they are happy with it.

{{endtip}}

For this and the next exercises, we'll use Watermill components.
It's up to you how you want to implement things in your project.
If you want, you can try to implement it from scratch.

{{tip}}

It's possible to use Watermill SQL Pub/Sub for forwarding even in a non-Go codebase.
You can write a small service (even up to 100 lines of code) that will be responsible for forwarding.
Then, you can write code in any language you want and store events in the database.

{{endtip}}

Let's start with publishing events in a database transaction.

## Exercise

Exercise path: ./11-outbox/05-publishing-events-in-transaction/main.go

**Implement the `PublishInTx` function:**

```go
func PublishInTx(
	message *message.Message,
	tx *sql.Tx,
	logger watermill.LoggerAdapter,
) error {
```

It should use the [SQL Pub/Sub](https://watermill.io/pubsubs/sql/).

1. Create a new SQL publisher.

```go
publisher, err := watermillSQL.NewPublisher(
    db,
    watermillSQL.PublisherConfig{
        SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
    },
    watermill.NewSlogLogger(nil),
)
```

Usually, `NewPublisher` accepts an `*sql.DB` as the first argument.
**Here, you want to use `tx` instead.**
It's how you create a publisher that works within the transaction.

Use `watermillSQL.DefaultPostgreSQLSchema{}` as the schema adapter.

{{tip}}

Initially, it may seem strange that we create a new publisher for each transaction.

However, if we want to publish an event within a transaction, the publisher needs to be _transaction-aware_.
It's not an issue in practice — it has no side effects and is fast enough.

{{endtip}}

2. Publish the provided message to the **`ItemAddedToCart` topic.**
