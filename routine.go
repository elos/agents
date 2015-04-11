package agents

import (
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/models"
)

type RoutineAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	models.Store
	models.User
	ticker *time.Ticker

	candidate models.Action
}

func NewRoutineAgent(store models.Store, u models.User) *RoutineAgent {
	return &RoutineAgent{
		Store:   store,
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
