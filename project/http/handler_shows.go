package http

import (
	"fmt"
	"net/http"
	ticketsEntity "tickets/entities"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/labstack/echo/v4"
)

type CreateShowRequest struct {
	DeadNationID    string    `json:"dead_nation_id"`
	NumberOfTickets int       `json:"number_of_tickets"`
	StartTime       time.Time `json:"start_time"`
	Title           string    `json:"title"`
	Venue           string    `json:"venue"`
}

func (h Handler) CreateShow(c echo.Context) error {
	var request CreateShowRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	show := ticketsEntity.Show{
		ShowID:          watermill.NewUUID(),
		DeadNationID:    request.DeadNationID,
		NumberOfTickets: request.NumberOfTickets,
		StartTime:       request.StartTime,
		Title:           request.Title,
		Venue:           request.Venue,
	}
	if err := h.showRepository.AddShow(c.Request().Context(), show); err != nil {
		return fmt.Errorf(err.Error())
	}
	data := map[string]string{"show_id": show.ShowID}

	return c.JSON(http.StatusCreated, data)
}
