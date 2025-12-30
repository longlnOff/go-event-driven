# Data Lakes

We {{exerciseLink "recently" "14-message-ordering" "06-project-remove-status-field"}} removed the status field from our read model.
It didn't require a migration because we had no old data in our read model.

But in the real world, it's not that simple.
We usually have old data in our read models and need to migrate them.

We need to prepare you to build event-driven systems in the wild.
And in the real world, we can't ignore the process of migrating read models.
Especially that removing a field is not the only case when we need to migrate read models.

There are also a couple other cases when migrating read models is needed:
- We need to add to our read model some data that we didn't have before.
- We had a bug in building our read model that we need to fix.
- We want to build a new read model for old events.

In all these cases, it would help a lot if we had stored all past events, so we could replay them.
The perfect pattern for this is a data lake.
The idea is simple: We should have a place where we store all events from our Pub/Sub.
It's a good idea to store them in a raw format, so we can easily replay them.
**Events are facts that happened in our system, so we should not modify them: They should be read-only.**

{{tip}}

Data lakes are also useful for analytics and reporting.
Company data teams can use events from the data lake to build reports and dashboards.
Data scientists can also use these events to build machine learning models.

Data in the data lake is read-only. This means even semi-technical product managers can query it to make decisions!

{{endtip}}

A data lake is one of those places where integrating through a database is not a bad idea.
It's entirely fine for all systems that need to read events from a data lake to use the same database.
Please bear in mind that this has a cost: Your events schema is now part of the contract.
You can't change it without ensuring that nobody depends on it. 
It could be an impossible task to get rid of an event at bigger organizations.
Fortunately, there are some strategies to overcome this issue that we will cover in the next module.

At bigger scale, it's important to have storage that allows storing a large volume of data at a low cost.
By larger scale, we are talking about terabytes or even petabytes of data.
When your scale is not big, it's not a challenge to store all events even in your PostgreSQL — that can be a good place to start.


{{tip}}

Choosing storage for your data lake is out of the scope of this training.
We will focus on how you can store events in a way that you can easily replay them.

Technologies that you can consider are:
- [Google Cloud BigQuery](https://cloud.google.com/bigquery)
- [Amazon Redshift](https://aws.amazon.com/redshift/)
- [ClickHouse](https://clickhouse.com/) _(open-source)_
- [Hadoop](https://hadoop.apache.org/) _(open-source)_
- PostgreSQL, MySQL, MongoDB, or Elasticsearch, if your event volume is not too big (like a few TB of data). _(open-source)_

**Using proprietary solutions like BigQuery or Redshift can lead to vendor lock-in.
On the other hand, hosting your own data lake at scale is not a simple task.
Choosing storage for a data lake is a technical decision.
It will have a big impact on your system.
Please research the topic thoroughly before making a decision.**

It's worth mentioning that we see a trend of shifting to cloud solutions instead of self-hosted ones.
Running your own Hadoop cluster is definitely not a simple task.
**Watch out for premature optimization:** It's easy to spend too much time setting up a complicated data lake at the beginning of a project.

It's hard to get advice that always works here. You need to use your own best judgement.
We have seen projects prepared for terabytes of data that ended with only a few gigabytes.

On the other hand, some projects prepared for a few gigabytes ended up with terabytes of data.
These projects required expensive migrations. But maybe the time that those projects saved
by not setting up a data lake helped them succeed?

If you are not expecting more than a terabyte of data soon, even a simple PostgreSQL would probably be enough.

{{endtip}}

It's time to add a data lake to your project.
