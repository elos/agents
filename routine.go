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

	data.Access
	models.User
	ticker *time.Ticker

	candidate models.Action
}

func NewRoutineAgent(a data.Access, u models.User) *RoutineAgent {
	return &RoutineAgent{
		Access:  a,
		User:    u,
		Life:    autonomous.NewLife(),
		Stopper: make(autonomous.Stopper),
	}
}

func (a *RoutineAgent) Start() {
	a.Life.Begin()
	<-a.Stopper
	a.Life.End()
}
