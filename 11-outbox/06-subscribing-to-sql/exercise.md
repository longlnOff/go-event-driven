# Subscribing to SQL

Now let's subscribe to the events in the SQL database.

We'll use the SQL subscriber from [`github.com/ThreeDotsLabs/watermill-sql/v3`](https://github.com/ThreeDotsLabs/watermill-sql).
This subscriber supports consumer groups.

It creates two tables. One is for storing messages:

```postgresql
CREATE TABLE IF NOT EXISTS watermill_topic (
    "offset" BIGSERIAL,
    "uuid" VARCHAR(36) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "payload" JSON DEFAULT NULL,
    "metadata" JSON DEFAULT NULL,
    "transaction_id" xid8 NOT NULL,
    PRIMARY KEY ("transaction_id", "offset")
);
```

The other one is for keeping offsets of the last processed message:

```postgresql
CREATE TABLE IF NOT EXISTS watermill_offsets_topic (
    consumer_group VARCHAR(255) NOT NULL,
    offset_acked BIGINT,
    last_processed_transaction_id xid8 NOT NULL,
    PRIMARY KEY(consumer_group)
);
```

{{tip}}

If you have a special use case, you can provide your own schema and offset adapters.

{{endtip}}

## Exercise

Exercise path: ./11-outbox/06-subscribing-to-sql/main.go

**Implement the `SubscribeToMessages` function.**

```go
func SubscribeToMessages(
    db *sqlx.DB,
    topic string,
    logger watermill.LoggerAdapter,
) (<-chan *message.Message, error) {
    // TODO: your code goes here
    return nil, nil
}
```

1. Create an SQL subscriber with `sql.NewSubscriber`.
Use the default configuration:

```go
subscriber, err := watermillSQL.NewSubscriber(
    db,
    watermillSQL.SubscriberConfig{
        SchemaAdapter:  watermillSQL.DefaultPostgreSQLSchema{},
        OffsetsAdapter: watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
    },
    logger,
)
```

**Important:** The Watermill SQL Pub/Sub requires schema initialization to work properly.

To initialize the schema, **you need an explicit call:**

```go
err := subscriber.SubscribeInitialize(topic)
```

If you don't call this function, you will receive an error like:

```text
could not insert message as row: pq: relation "watermill_ItemAddedToCart" does not exist
```

Alternatively, you can set the `InitializeSchema` config option to `true` — it does the same thing.

2. After creating and initializing the subscriber, call `subscriber.Subscribe()` and return the messages as the result.
