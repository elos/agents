package agents

import (
	"log"
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
	"github.com/elos/models/action"
)

type ActionAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	*data.Access
	models.User
	ticker *time.Ticker
}

func NewActionAgent(a *data.Access, u models.User) *ActionAgent {
	return &ActionAgent{
		Access:  a,
		User:    u,
		Life:    autonomous.NewLife(),
		Stopper: make(autonomous.Stopper),
	}
}

func (a *ActionAgent) Start() {
	a.ticker = time.NewTicker(1 * time.Second)
	a.Life.Begin()
	log.Print("Action Agent Booting up")

	// FIXME subscribe directly to user, db access can be different
	// than the user
	changes := *a.Access.RegisterForChanges()

Run:
	for {
		select {
		case c := <-changes:
			go a.changeSieve(c)
		case <-a.ticker.C:
			a.TryNewAction()
		case <-a.Stopper:
			break Run
		}
	}

	a.User.SetCurrentAction(nil)
	a.Life.End()
}

func (a *ActionAgent) changeSieve(c *data.Change) {
	if c.Record.ID() != a.Client.ID() {
		return
	}

	a.TryNewAction()
}

func (a *ActionAgent) TryNewAction() {
	a.PopulateByID(a.User) // reload

	act, _ := action.New(a.Store)

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
