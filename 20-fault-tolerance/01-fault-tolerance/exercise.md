# Fault Tolerance

Over the previous few modules, we described a few patterns that help with error handling.
In this module, we dive deeper into ideas related to fault tolerance.

**In your production systems, errors will happen.**
Even if you write perfect code — and you won't — you can't control factors outside it.
You will subscribe to events published by another party, and you will call external services or interact with databases.
All of these can fail, and you need to be prepared for that.

We recommend looking for patterns that allow the system to auto-heal with no human intervention.
This is especially important in distributed systems, where a single failure can cause a cascade of issues.
You can use a decoupled microservices architecture, but if failure of one of the services causes others to break,
your system is far from resilient.

Let's look into ways to prevent this.
