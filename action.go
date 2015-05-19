package agents

import (
	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
)

type (
	ActionAgent struct {
		autonomous.Life
		autonomous.Managed
		autonomous.Stopper

		data.DB
		*models.User

		man autonomous.Manager
		*RoutineAgent
		*CalendarAgent
	}
)

func NewActionAgent(db data.DB, u *models.User) *ActionAgent {
	h := autonomous.NewHub()
	go h.Start()

	ra := NewRoutineAgent(db, u)

	return &ActionAgent{
		Life:    autonomous.NewLife(),
		Stopper: make(autonomous.Stopper),

		DB:   db,
		User: u,

		man:          h,
		RoutineAgent: ra,
	}
}

func (a *ActionAgent) Start() {
	a.Life.Begin()

	changes := data.Filter(data.FilterKind(a.DB.Changes(), models.UserKind), func(change *data.Change) bool {
		return change.Record.ID() == a.User.ID()
	})

Run:
	for {
		select {
		case <-*changes:
			a.tryNewAction()
		case <-a.Stopper:
			break Run
		}
	}

	a.Life.End()
}

func (a *ActionAgent) tryNewAction() {
	if err := a.DB.PopulateByID(a.User); err != nil {
		return
	}

	action, err := a.User.CurrentAction(a.DB)
	if err != nil {
		return
	}

	if !action.Completed {
		return
	}
}
