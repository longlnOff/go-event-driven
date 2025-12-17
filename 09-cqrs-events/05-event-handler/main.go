package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type FollowRequestSent struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type EventsCounter interface {
	CountEvent() error
}

type RequestHandler struct {
	EventsCounterService EventsCounter
}

func (r *RequestHandler) HandleRequest(ctx context.Context, event *FollowRequestSent) error {
	return r.EventsCounterService.CountEvent()

}

func NewFollowRequestSentHandler(counter EventsCounter) cqrs.EventHandler {
	RequestHandler := RequestHandler{
		EventsCounterService: counter,
	}
	return cqrs.NewEventHandler(
		"FollowRequestSentHandler",
		RequestHandler.HandleRequest,
	)
}
