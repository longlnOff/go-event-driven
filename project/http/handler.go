package http

import (
	ticketsWorker "tickets/worker"
)

type Handler struct {
	worker *ticketsWorker.Worker
}
