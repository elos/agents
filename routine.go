package agents

import (
	"log"
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
	"github.com/elos/models/action"
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
	a.ticker = time.NewTicker(1 * time.Second)
	a.Life.Begin()

	// FIXME subscribe directly to user, db access have
	// a client different than theusercan be different than the user
	changes := *a.Access.RegisterForChanges(a.User)

Run:
	for {
		select {
		case c := <-changes:
			go a.changeSieve(c)
		case <-a.ticker.C:
			a.check()
		case <-a.Stopper:
			break Run
		}
	}

	a.User.SetCurrentAction(nil)
	a.Life.End()
}

func (a *RoutineAgent) changeSieve(c *data.Change) {
	if c.Record.ID() != a.Client().ID() {
		return
	}

	a.check()
}

func (a *RoutineAgent) ActionCandidate() models.Action {
	return a.candidate
}

func (a *RoutineAgent) check() {
	a.PopulateByID(a.User) // reload

	act, _ := action.New(a)

	if err := a.User.CurrentAction(a.Access, act); err != nil {
		log.Print(err.Error())
	}

	if !act.Completed() {
		return // not done
	}

	actionable, err := a.User.CurrentActionable(a.Access)
	if err == data.ErrNotFound || err != nil {
		return // we got nothin to do
	}

	actionable.CompleteAction(a.Access, act)

	nextAction, ok := actionable.NextAction(a.Access)
	if !ok {
		return
	}

	a.User.SetCurrentAction(nextAction)
	nextAction.SetStartTime(time.Now())

	a.Save(nextAction)
	a.Save(a.User)
	log.Print("Action Agent Set New Action")
}
