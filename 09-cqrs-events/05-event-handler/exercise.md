# Implementing the Event Handler

Now let's implement the Event Handler.
You can use the `NewEventHandler` function for this.

```go
cqrs.NewEventHandler(
	"ArticlePublishedHandler", 
	// This function will subscribe to the topic where ArticlePublished events are published.
	func(ctx context.Context, event *ArticlePublished) error {
		fmt.Printf("Article %s published\n", event.ArticleID)
		
		return nil
	},
),
```


#### Injecting dependencies into handler from `NewEventHandler`

You can use the same technique that you likely already use with HTTP handlers to inject dependencies into the event handler methods.
You can create a struct that holds all dependencies, and then pass the method as a value to `cqrs.NewEventHandler`:

```go
type ArticlesHandler struct {
	notificationsService NotificationsService
}

func (h ArticlesHandler) PrintIDOnArticlePublished(ctx context.Context, event *ArticlePublished) error {
	fmt.Printf("Article %s published\n", event.ArticleID)
	
	return nil
}

func (h ArticlesHandler) NotifyUserOnArticlePublished(ctx context.Context, event *ArticlePublished) error {
	h.notificationsService.NotifyUser(event.ArticleID)
	
	return nil
}

func NewArticlesHandlers(notificationsService NotificationsService) []cqrs.EventHandler {
	h := ArticlesHandler{
		notificationsService: notificationsService,
	}

	return []cqrs.EventHandler{
		cqrs.NewEventHandler(
			"PrintIDOnArticlePublished", 
			h.PrintIDOnArticlePublished,
		), 
		cqrs.NewEventHandler(
			"NotifyUserOnArticlePublished", 
			h.NotifyUserOnArticlePublished, 
		),
	}
}
```

Note that the `ArticlesHandler` can have multiple handlers for the same event.

## Exercise

Exercise path: ./09-cqrs-events/05-event-handler/main.go

Implement the `NewFollowRequestSentHandler` function that returns `cqrs.EventHandler`.
It should accept `EventsCounter` as a parameter.

`CountEvent()` should be called each time the event is handled.
