# Replacing Dead Nation

{{background}}

Our company grows!
We sell more and more tickets, and we start to hit the wall of the current setup.

Currently, we use the "Dead Nation" company to handle all the logic related to ticket reservation, keeping track of the seat limits, etc.
For example, the well-known `POST /tickets-status` endpoint is called by Dead Nation when a ticket is confirmed.

Dead Nation starts to become a problem for multiple reasons.
First of all, they don't support reserving individual seats, so we can't support big events that happen in stadiums.

Sometimes their system also doesn't work properly — for example, they might allow booking a ticket for a show that is sold out.
It's problematic to handle some exclusive events with big demand: Imagine selling 500 tickets for show limited to 50 people.

These issues make our customers angry and generate a ton of work for our ops team.

Sometimes it's even worse: Their system goes down for a long time, especially if we have a big event.

There is also one more problem: They aren't cheap.
Dead Nation takes a significant cut from our margin.

It's time to consider moving out of Dead Nation.
It would be a big engineering effort on our end, but we have a good (financial) reason: All of the previously described problems cost us a lot of money.

Big migrations are always risky.
We shouldn't make this migration a big bang event but take it step by step.

{{endbackground}}

## The Strangler Pattern

To migrate from Dead Nation, we will use a form of the [Strangler Pattern](https://www.martinfowler.com/bliki/StranglerFigApplication.html).

In the Strangler Pattern, we create a "Strangler" part of our application that is a facade in our system and proxies all incoming requests.
Later, the "Strangler" part delegates all requests both to new and old parts of our application.

The legacy system works as before, but in the meantime, we can migrate all functionality to the new system.
Once the new application is ready, we can do cross-checks between systems to ensure that none of the data was lost and everything works as expected.

This is how this usually works at a high level:

```plantuml
@startuml

skinparam monochrome true

!define ICONURL https://raw.githubusercontent.com/RicardoNiepel/C4-PlantUML/v2.5.0/dist

!includeurl ICONURL/C4.puml

title Strangler Pattern

actor "User" as user

node "Legacy System" as legacy {
  component "Legacy Application" as legacyApp
}

node "Strangler Application" as strangler {
  database "Strangler Database" as stranglerDB
  component "Strangler Module" as stranglerModule
}

node "New Application" as newApp {
  database "New Database" as newDB
  component "New Module" as newModule
}

user -> stranglerModule: Use Strangler Features

stranglerModule --> legacyApp: Invoke legacy functionality
stranglerModule --> newModule: Delegate requests

stranglerModule --> stranglerDB: Read/Write Data
newModule --> newDB: Read/Write Data

@enduml
```

{{tip}}

In many cases, it's good to keep the outputs and behaviors of both systems identical at the beginning — 
this makes it easier to compare results between the old and new versions to ensure that the new one works properly.

Try to get rid of the legacy part of the application ASAP and implement new features in new system after it's deployed.
Don't rewrite the entire application at once — find small parts that can be migrated, and do it step by step.

The alternative is a never-ending story of having "the new system" that will solve everything but is in development for years.
It's very, very common for this to happen: Avoid this at all costs because it never ends well.

{{endtip}}

The integration can be done via synchronous calls, but that may be risky when integrating with legacy systems.
If you don't need a synchronous response to the public API from your legacy system, it's a good idea to integrate over Pub/Sub.

The previous diagram is, of course, just a general idea.
It rarely looks exactly like that in real life — it's about illustrating the high-level concept.
You should adjust it to your needs.

{{tip}}

It's not a requirement that components from the diagram be separate services/microservices.
The only requirement is good modularization, but that can be achieved in a single monolith if it works best for you.

{{endtip}}

This is how it will look like in our case:


```plantuml
@startuml

skinparam monochrome true

!define ICONURL https://raw.githubusercontent.com/RicardoNiepel/C4-PlantUML/v2.5.0/dist

!includeurl ICONURL/C4.puml

title Strangler Pattern

actor "User" as user

node "Dead Nation's\nPOST /book-tickets" as legacy {
  component "Legacy Application" as legacyApp
}

node "POST /book-tickets" as strangler {
  database "Database" as stranglerDB
  component "HTTP handler" as stranglerModule
}

node "POST /tickets-status" as newApp {
  database "Database" as newDB
  component "HTTP Handler" as newModule
}

user -> stranglerModule: Use Strangler Features

stranglerModule --> legacyApp: Invoke legacy functionality
legacyApp ->> newModule: Delegate requests
stranglerModule ..> newModule: Call directly (after removing Dead Nation)

stranglerModule --> stranglerDB: Read/Write Data
newModule --> newDB: Read/Write Data

@enduml
```

We will introduce a new endpoint: `POST /book-tickets`.
This is the endpoint called by our frontend when the user books a ticket.
Now it's handled by Dead Nation — we want to intercept this request and let Dead Nation call `POST /tickets-status`.
After replacing Dead Nation, we will call `POST /tickets-status` directly.

The high-level idea of the Strangler Pattern is maintained: We have a part of our application that catches all requests and delegates them to the legacy system.
The difference, in our scenario, is that we won't remove this component after migration: It will just be a new endpoint.

That's the high-level idea. In the next few exercises, we'll do a step-by-step migration.
