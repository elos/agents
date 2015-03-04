package agents

import (
	"log"
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
)

type ActionAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	data.Access
	models.User

	*autonomous.Hub
	*RoutineAgent
	*CalendarAgent
}

func NewActionAgent(a data.Access, u models.User) *ActionAgent {
	return &ActionAgent{
		Access:  a,
		User:    u,
		Life:    autonomous.NewLife(),
		Stopper: make(autonomous.Stopper),
		Hub:     autonomous.NewHub(),
	}
}

func (a *ActionAgent) Start() {
	// This agent is a hub, itself.
	go a.Hub.Start()

	// Routine Agent is currently the only known actionable
	// entity.
	a.RoutineAgent = NewRoutineAgent(a.Access, a.User)
	go a.StartAgent(a.RoutineAgent)

	changes := *a.Access.RegisterForChanges(a.User)
	a.Life.Begin()

Run:
	for {
		select {
		case c := <-changes:
			go a.changeSieve(c)
		case <-a.Stopper:
			break Run
		}
	}

	a.User.SetCurrentAction(nil)
	a.Life.End()
}

func (a *ActionAgent) changeSieve(c *data.Change) {
	if c.Record.ID() == a.Client().ID() { // The user changed
		a.TryNewAction()
	}
}

func (a *ActionAgent) reload() {
	a.PopulateByID(a.User)
}

func (a *ActionAgent) TryNewAction() {
	act, err := a.User.CurrentAction(a.Access)
	if err != nil {
		log.Printf("TODO %s", err.Error())
	}

	if !act.Completed() {
		return // not done
	}

	actionable, err := a.User.CurrentActionable(a.Access)
	if err == data.ErrNotFound {
		a.Delegate()
		// check with routine
		return
	}

	if err != nil {
		a.Delegate()
		return // shit
	}

	a.CompleteAction(act, actionable)
}

func (a *ActionAgent) CompleteAction(act models.Action, actionable models.Actionable) {
	actionable.CompleteAction(a.Access, act)

	nextAction, err := actionable.NextAction(a.Access)

	if err != nil {
		a.DemoteCurrentActionable()
		a.Delegate()
		return
	}

	a.User.SetCurrentAction(nextAction)
	nextAction.SetStartTime(time.Now())

	a.Save(nextAction)
	a.Save(a.User)

	log.Print("Action Agent Set New Action")
}

func (a *ActionAgent) DemoteCurrentActionable() error {
	a.User.ClearCurrentActionable()
	return a.Save(a.User)
}

// Invariants: the current action is complete and
// the current actionable is null
func (a *ActionAgent) Delegate() {
	cand, exists := a.CalendarAgent.Candidate()
	if exists {
		act, err := a.CalendarAgent.ResponsibleActionable()
		if err != nil {
			a.SetCurrentAction(cand)
			a.SetCurrentActionable(act)
		}
	}
	// ask calendar
	return
}

func (a *ActionAgent) SetCurrentAction(act models.Action) {
	a.User.SetCurrentAction(act)
	a.Access.Save(a.User)
}

func (a *ActionAgent) SetCurrentActionable(act models.Actionable) {
	a.User.SetCurrentActionable(act)
	a.Access.Save(a.User)
}
