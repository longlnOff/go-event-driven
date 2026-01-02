# Pub/Sub metrics

As mentioned earlier, some Pub/Subs provide certain metrics out of the box.

One alternative way of getting metrics is using one of the pre-built Prometheus exporters for message brokers.
It's unlikely that they would provide you detailed metrics about message processing, but they can provide you some extra high-level metrics like the following:

- The number of messages in the topic
- The oldest message in the topic
- The number of messages in the subscription (or consumer group)
- The oldest message in the subscription (or consumer group)

{{img "Oldest unacked messages" "gcp-oldest-unacked.png"}}

{{img "Unacked messages" "gcp-unacked-count.png"}}

You can check the [list of exporters on the Prometheus website](https://prometheus.io/docs/instrumenting/exporters/#messaging-systems).

If you don't find the right exporter for your Pub/Sub, you can always write it by yourself.
It's not that hard.

This may be something you may want to do for monitoring queue sizes in the {{exerciseLink "outbox" "11-outbox" "06-subscribing-to-sql"}}.
Thanks to that, if messages from the outbox become stuck in the queue for any reason, you will be able to notice it.
Often, when those messages are stuck, it may mean that the system is frozen due to a lack of all events sent via outbox.

{{tip}}

We won't have an exercise to export the size of the outbox to Prometheus, but you should be able to do it by yourself.
We highly encourage you to do so.

{{endtip}}
