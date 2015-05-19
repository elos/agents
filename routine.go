package agents

import (
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
)

type RoutineAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	data.DB
	*models.User
	ticker *time.Ticker

	candidate models.Action
}

func NewRoutineAgent(db data.DB, u *models.User) *RoutineAgent {
	return &RoutineAgent{
		Life:    autonomous.NewLife(),
		Stopper: make(autonomous.Stopper),

		DB:   db,
		User: u,
	}
}

func (a *RoutineAgent) Start() {
	a.Life.Begin()
	<-a.Stopper
	a.Life.End()
}
