# Failed to Book Taxi

The last piece of the puzzle is to handle taxi booking failure.

```mermaid
stateDiagram-v2
    [*] --> VipBundleInitialized
    VipBundleInitialized --> BookShowTickets : OnVipBundleInitialized
    BookShowTickets --> BookingMade : Success
    BookShowTickets --> BookingFailed : Failure
    BookingMade --> BookInboundFlight : OnBookingMade
    BookInboundFlight --> FlightBooked : Success
    BookInboundFlight --> FlightBookingFailed : Failure
    FlightBooked --> BookReturnFlight : InboundBooked & ReturnNotBooked
    FlightBooked --> BookTaxi : Both Flights Booked
    BookReturnFlight --> FlightBooked : Success
    BookReturnFlight --> FlightBookingFailed : Failure
    FlightBookingFailed --> Rollback : OnFlightBookingFailed
    BookTaxi --> TaxiBooked : Success
    BookTaxi --> TaxiBookingFailed : Failure
    TaxiBooked --> VipBundleFinalized : OnTaxiBooked
    TaxiBookingFailed --> Rollback : OnTaxiBookingFailed
    Rollback --> RefundTicket : Show Booking Exists
    Rollback --> CancelInboundFlight : Inbound Flight Exists
    Rollback --> CancelReturnFlight : Return Flight Exists
    Rollback --> FailedState
    VipBundleFinalized --> FinalizedState
    FailedState --> [*] : Failed Process
    FinalizedState --> [*] : Completed Process
```

As with the flight bookings, the taxi service will now return HTTP 409.

## Exercise

Exercise path: ./project

All that we need to do here is to publish the `TaxiBookingFailed_v1` event.
It should be handled by our process manager, and all operations should be rolled back.

You should already have all command and event handlers to handle it.

