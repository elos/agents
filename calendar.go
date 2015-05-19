package agents

import (
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
)

type CalendarAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	data.DB
	*models.Calendar
	*models.User

	ticker *time.Ticker

	candidate models.Action
}

func NewCalendarAgent(db data.DB, u *models.User) *CalendarAgent {
	c, _ := u.Calendar(db)

	return &CalendarAgent{
		DB:       db,
		User:     u,
		Calendar: c,
		Life:     autonomous.NewLife(),
		Stopper:  make(autonomous.Stopper),
	}
}
