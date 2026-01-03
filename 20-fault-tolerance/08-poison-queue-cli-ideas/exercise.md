# Poison Queue CLI: Ideas

The CLI is already a solid starting point but far from complete.
We share a few ideas below that can come in handy when building your own tooling.

### Filter Messages by UUID or Correlation ID

Sometimes, you know exactly what message you're looking for.
You can add flags to display or target only messages with the provided UUID, 
or, with the provided Correlation ID, kept in either the event header or the message metadata.
It's easy to get both IDs from the application logs, metrics, or tracing.

### Filter Messages from a Given Topic or Handler 

The poison queue middleware adds the original topic and handler as metadata to the message.
You can use it to filter messages from a given topic or handler.

### Allow Modifying Messages In-Place 

If a message is somehow malformed, you may want to edit it and republish it to the original topic.
The CLI could allow you to do this, which makes the most sense if you use payloads in a human-readable format like JSON.

### Add Locks 

Make sure multiple people using the CLI don't interfere with each other.
You can use some kind of remote locks to allow only one person to work on a message at a time.

### Add a Web UI

If you want to go the extra mile, you can build a simple Web UI for managing the poison queue on top of the existing features.
You can use the CLI code as a library and build a simple HTTP server on top of it.
