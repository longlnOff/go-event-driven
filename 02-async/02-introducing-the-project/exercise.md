# Introducing the Project

It's time to kick off your long-running project. 

{{background}}

**You've just joined a ticket aggregator startup as a senior engineer.**

The company integrates with a few ticket vendors, so users can buy all their tickets in one place.

The good news is that the company has been doing very well over the last couple of months.
The product-market fit is there, and many clients are signing up every day.

Sadly, the codebase is a mess.
The MVP was successful at getting VCs to invest, but the naive architecture can't handle the load anymore.
You've been hired to fix this!

{{endbackground}}

### Naming conventions

We use the following naming conventions in this project:

- **API** - an external service that we call.
- **Service** - a part of our platform that we call.

### The Common Package

We want you to focus on the event-driven part of the project.
We prepared a `common` package (in a [public repository](https://github.com/ThreeDotsLabs/go-event-driven))
with ready-to-use code for things like the HTTP server, HTTP clients, and logging.
You don't need to use it, but it's there if you want to.

### The Gateway

All project exercises (and some non-project as well) use external APIs via HTTP.
They're all available through a *gateway service* (a reverse proxy).
You can access it by the `GATEWAY_ADDR` environment variable.
You'll also be able to run it locally.

### Project Structure

There is no perfect way of organizing code.

You don't need to follow exactly the same layout as we do in the example solution.
We used a format that should be easy to understand by anyone.

**You don't need to replicate this layout in your project. Use what works best for you.**
If you keep the same structure, it will be easier to follow the example solutions,
and especially the diffs between the exercises.

If you want to use your own layout, that's perfectly fine!
As long as your code passes the tests, you're good to go.
Just keep in mind that the example solutions won't be as helpful!

We start with three packages:

* **`adapters`** — contains adapters to infrastructure and external APIs.
* **`http`** — contains the HTTP server and handlers.
* **`service`** — contains the bootstrapping code.

{{tip}}

We use a lightweight layered architecture, similar to *clean architecture*.
**The TL;DR is that we separate the code in packages with different responsibilities.**

We start like this because the project will grow over time, and we want to keep it maintainable.

You don't need to be familiar with it to complete the exercises.
You will be fine just following along with the code.

If you want to read more, check our blog: [How to implement Clean Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/).

{{endtip}}
