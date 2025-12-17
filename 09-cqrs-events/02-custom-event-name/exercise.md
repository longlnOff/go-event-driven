# Custom Event Name

If you check the logs, you'll notice that `DocumentPrinted` is published to the `main.DocumentPrinted` topic — it doesn't look great. 
By default, the JSON marshaler uses the full name of the event struct as the event name.

Let's change this.

## Exercise

Exercise path: ./09-cqrs-events/02-custom-event-name/main.go

The `JSONMarshaler` has an attribute `GenerateName`. Let's change it to `cqrs.StructName`.
This will generate the event name as `DocumentPrinted` instead of `main.DocumentPrinted`.

{{tip}}

Remember to pass the function to the `GenerateName` of `JSONMarshaler`.
Be careful to not pass the result of the function call!
This field accepts a function.

{{endtip}}
