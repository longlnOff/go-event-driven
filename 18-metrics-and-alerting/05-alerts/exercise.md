# Alerts

It's now clear how we can export all the needed data to Prometheus.
However, having just stored data is not enough — we need to have a way to know when our system is not healthy.
Doing this shouldn't require having a dedicated screen that we're constantly scanning to see if everything is fine —
let's leave that for the basement-dwelling hackers in movies.
Instead, we need a way to be notified when something goes wrong.

We won't go into the details of implementing alerts in this training because it will be far beyond the scope.

In this exercise, we'll give you a high-level overview of what kinds of alerts you should have in your system and how to define them.

{{tip}}

In the table, we use _consumer group_ to show the grouping mechanism.
This may be _subscription_ or _consumer group_, depending on the Pub/Sub you are using.
What's important is that this thing is supposed to be the grouping mechanism for your consumers.
It's important to not set alerts to the topic but to the _grouping thing_ instead because that will give you more context about the incident.
There may be two different consumers on the same topic with totally different failure severities.

For example, let's imagine an event like `FlightCancelled` for an airline company,
which may handle some critical process like notifying the customer about a cancellation.
Any message that couldn't be processed should be treated as critical; if it happens in the middle of the night, someone should be woken up.

However, if this event will be handled by another consumer that is responsible for building some non-important read model or report, it's not that critical.
It can wait until morning to be resolved.

Marking whether a topic is critical or not can be done by adding a label to the metric related to specific subscriptions.

It may also be a good idea to add a label with the ownership of the subscription. 
This may make it easier to route the alert to the right team.
It's a mistake when you receive alerts to topics that you are just emitting to (but you are not consuming from them). 
Of course, that's assuming you are not emitting malformed messages.

{{endtip}}

{{img "Example Chart" "gcp-per-subscribtion.png"}}

| Alert name                                  | Group by                      | Description                                                                                                                                                                 | Critical                                                                          | Threshold                                                                                                                                       |
|---------------------------------------------|-------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| Unprocessed messages on the topic           | Consumer group, topic         | If above zero, it may mean that your handlers are not fast enough to process messages or there is a {{exerciseLink "spinning message" "07-errors" "04-project-code-bugs"}}. | Depends on count and group                                                        | Depends on the scale of your system and how it's counted; 1 may be too extreme                                                                  |
| Unprocessed messages in outbox              | Topic                         | If above zero, a {{exerciseLink "process" "11-outbox" "08-publishing-events-with-forwarder"}} responsible for forwarding your messages may not work properly.                 | Forwarding from outbox should be fast and may block many processes in your system | Depends on the scale of your system and how it's counted; 1 may be too extreme                                                                  |
| Oldest message on the topic                 | Consumer group, topic         | It's a different metric that can show if some message is spinning.                                                                                                          | Depends on group                                                                  | Depends on the SLA of your system, but generally 1-5 minutes                                                                                    |
| Message processing duration                 | Topic, handler name, quantile | Higher values show performance issues related to message processing.                                                                                                        | No, as long as you don't have a specific SLA for processing messages              | Depends on your system, but usually 10-60s                                                                                                      |
| Latency between message publish and consume | Topic, handler name / group   | Higher values than usual shows performance issues of your messaging infrastructure or performance issues related to message processing.                                     | No, as long as you don't have a specific SLA for processing messages              |                                                                                                                                                 |
| Message processing error rate               | Topic, handler name / group   | A higher rate shows that more transient failures are happening (may be not visible in other metrics if messages are processed after retries).                               | Depends on the handler/group and how often processing fails                       | Depends on our system                                                                                                                           |
| Message processing rate                     | Topic, handler name / group   | Values that are higher than normal can show anomalies in the system (may be a DDoS or bug).                                                                                 | No, as long as the processing rate is not much higher than usual (10-100x)        | Should be based on a normal processing rate in your system (you can have non-critical alert for 3x more messages and critical for 10x messages) |


Those are just examples of alerts that you may want to have in your system.
They should serve as inspiration for you, which you should adapt to your needs.

We didn't provide specific values for thresholds because they are very specific to your system.
You should start to gather metrics, find some baseline, and then set alerts based on that.
It's usually a good idea to start with more aggressive metrics and then relax them based on how your system is behaving.
It will be noisy at the beginning, but it's better than missing some incidents because of thresholds that are too relaxed.

Some of the alerts may overlap with each other.
In practice, the bare minimum is having any alert that will fire when a message is stuck on the topic.

If you don't have any tool for alerts yet, you can start with looking at [Prometheus Alertmanager](https://prometheus.io/docs/alerting/alertmanager/).
However, it's not a must to use it — any non-terrible alerting tool should integrate with Prometheus.

## Handling incidents

Having alerts is one thing, but you also need a process for handling incidents.
If you are the person who is adding alerts, you need to be sure that you don't end up as the only person who is solving them.

Having alerts requires a proper process in your team or organization.
Usually, you would like to define an on-call rotation in which all engineers are included.

In some cases, it could make sense to have two on-call rotations: one for business hours and one for off-hours.
This means that a person who is on call during the night won't be tired during the day.

Ideally, your team should be responsible for alerts for your services (in the spirit of _you build it, you run it_).
There's nothing worse than fighting with the alerts of a team that you don't know anything 
about and that doesn't care about improving the stability of the system.
Unfortunately, this is not always possible if your team is small (because you'd be on off-hours rotation very often).

If you're looking for a tool for handling alerts, consider [Grafana OnCall](https://grafana.com/products/cloud/oncall/) or for more advanced functionalities [PagerDuty](https://www.pagerduty.com/).
Both support defining an on-call rotation, escalation policies, and more.
When an incident happens, they can notify the person on call using the provided channel (SMS, application, phone call, etc.).
