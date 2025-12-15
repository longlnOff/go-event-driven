# Component Testing

In this module, we're not introducing any new functionality.
We'll focus on writing component tests for the existing code.

{{tip}}

This module is optional. If you want, you can skip it with `tdl tr skip`.

{{endtip}}

### Test Names Definition

Usually, the definition of _unit_ or _end-to-end (E2E) tests_ is clear for everyone, 
but we've seen a lot of misunderstanding about what _integration tests_ are.
It's even worse with _component tests_.

To start with, we need to be sure that we are on the same page with the definition of each type of test.

| Feature / Test Type           | Unit                       | Integration  | Component        | End-to-End |
|-------------------------------|----------------------------|--------------|------------------|------------|
| **Docker database**           | No                         | Yes          | Yes              | Yes        |
| **Use of external systems**   | No                         | No           | No               | Yes        |
| **Focused on business cases** | Depends on the tested code | No           | Yes              | Yes        |
| **Uses mocks or stubs**       | Most dependencies          | Usually none | External systems | None       |
| **Tested API**                | Go package                 | Go package   | HTTP and gRPC    | HTTP       |
| **Execution speed**           | Fast                       | Fast         | Medium           | Slow       |
| **Cost of introduction**      | \$                         | \$\$         | \$\$\$           | \$         |
| **Cost of maintenance**       | \$                         | \$\$         | \$\$             | \$\$\$\$   |

### Testing Strategies

Projects often start with a lot of unit and E2E tests because they are easy to introduce.
For very simple projects, relying on those tests may be enough. 
However, as the project grows in complexity, it becomes harder to maintain them.
Adding new features is also not easy, because E2E tests are slow and unit tests can't cover all features.
E2E tests are also usually harder to run locally, resulting in a longer development feedback loop.
It's a big hit to your productivity if you can't verify that a feature you implemented works locally.

Component tests are the missing link between unit and E2E tests.
In component tests, we test the integration between components of our application with a Docker database and external dependencies stubbed out.
Thanks to this, component tests are faster than E2E tests and cover more than unit tests.
It's also easier to test edge cases for external dependencies — it's easier to achieve the expected behaviour of external systems.

### How to write component tests?

Component tests have some similarities to E2E tests.
For component tests, we need to start the entire application and test it by using the public API.
However, we want to use stubs for external dependencies that we're unable to run in Docker.
In the case of our application, we will stub all APIs we're reaching through the Gateway.
As much as possible, we want to use real infrastructure in Docker.

This is what our infrastructure looks like in production.
(We are using just Redis, but it's likely that we'll need PostgreSQL in the future.)

```c4plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

title System Context diagram for Ticket Booking System

Person(user, "User", "A user who interacts with the system to book, confirm, or cancel tickets.")

Boundary(b, "Our infrastructure") {
  System(ticket_booking_system, "Ticket Booking System", "Handles ticket booking, confirmation, and cancellation events.")
  Container(gateway, "API Gateway", "Provides access to various services.")

  Container(redis, "Redis", "Used for message publishing and subscription.")
  ContainerDb(postgres, "PostgreSQL", "Persistent storage for the system.")
}

System_Ext(receipts_service, "Receipts Service", "Handles the creation of receipts.")
System_Ext(spreadsheets_service, "Spreadsheets Service", "Manages spreadsheets for printing tickets and refunds.")

Rel(user, ticket_booking_system, "Uses", "HTTP call")
Rel(ticket_booking_system, gateway, "Sends requests to")
Rel_Back(gateway, receipts_service, "Routes requests to")
Rel_Back(gateway, spreadsheets_service, "Routes requests to")
Rel(ticket_booking_system, redis, "Publishes and subscribes to events from")
Rel(ticket_booking_system, postgres, "Stores data in")

@enduml
```

_(Diagram generated from [C4-PlantUML](https://github.com/plantuml-stdlib/C4-PlantUML). 
Some time ago, our friend Krzysztof wrote a guest article on our blog on how to [generate C4 diagrams directly from Go code](https://threedots.tech/post/auto-generated-c4-architecture-diagrams-in-go/))._

To run component tests locally, we need to replace the external dependencies with stubs and run the required infrastructure locally.

```c4plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

title System Context diagram for Ticket Booking System

Person(tests, "Tests", "Automated tests interact with the system to validate its functionality.")

  System(ticket_booking_system, "Ticket Booking System", "Handles ticket booking, confirmation, and cancellation events.")
  
  Boundary(docker, "Docker-Compose") {
    Container(redis, "Redis", "Used for message publishing and subscription.")
    ContainerDb(postgres, "PostgreSQL", "Persistent storage for the system.")
  }

System_Ext(receipts_service_stub, "Receipts Service Stub", "Emulates the functionality of the Receipts Service.")
System_Ext(spreadsheets_service_stub, "Spreadsheets Service Stub", "Emulates the functionality of the Spreadsheets Service.")


Rel(tests, ticket_booking_system, "Uses", "HTTP call")
Rel(ticket_booking_system, receipts_service_stub, "Sends requests to")
Rel(ticket_booking_system, spreadsheets_service_stub, "Sends requests to")
Rel(ticket_booking_system, redis, "Publishes and subscribes to events from")
Rel(ticket_booking_system, postgres, "Stores data in")

@enduml
```

### Which Features to Test in Component Tests?

For most projects that are similar to our application, the best strategy is to test the happy paths of each added feature in component tests. 
Edge cases and unhappy paths should be tested in unit tests. 
The most critical scenarios should also be tested with both E2E tests and component tests.

For example, the happy paths for features in our ticket application could be the following:
- A receipt is issued for a confirmed ticket.
- The `tickets-to-print` sheet is updated with a new row when a ticket is confirmed.
- The `tickets-to-refund` sheet is updated with a new row when a ticket is canceled.

Component tests should ideally be written before enabling the feature in production.

{{tip}}

We deliberately use the word "enabling" and not "deploying." We encourage you to use [trunk-based development](https://trunkbaseddevelopment.com) and continuously deploy to production.

{{endtip}}

Now, let's write the component tests for our ticket application!

{{tip}}

Some companies use dedicated QA teams for writing most of their tests.
We can't know all types of companies, but we believe that it's better to have the team of developers be responsible 
for writing most of the tests. 
This helps decrease the cycle time for developing the feature and gives more ownership to the team.

{{endtip}}
