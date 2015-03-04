package agents

import (
	"errors"
	"log"
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
)

type CalendarAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	data.Access
	models.Calendar
	models.User

	ticker *time.Ticker

	candidate models.Action
}

func NewCalendarAgent(a data.Access, u models.User) *CalendarAgent {
	c, _ := u.Calendar(a)

	return &CalendarAgent{
		Access:   a,
		User:     u,
		Calendar: c,
		Life:     autonomous.NewLife(),
		Stopper:  make(autonomous.Stopper),
	}
}

func (a *CalendarAgent) Start() {
	a.Life.Begin()

	a.ticker = time.NewTicker(1 * time.Second)

Run:
	for {
		select {
		case <-a.ticker.C:
			a.check()
		case <-a.Stopper:
			break Run
		}
	}
	<-a.Stopper
	a.Life.End()
}

func (a *CalendarAgent) check() {
	nextFixture, err := a.Calendar.NextFixture(a.Access)
	if err != nil {
		return
	}

	if nextFixture.StartTime().Sub(time.Now()) < 1*time.Minute {
		a.placeCandidate(nextFixture)
	}
}

func (a *CalendarAgent) placeCandidate(f models.Fixture) {
	a.Calendar.SetCurrentFixture(f)
	a.Save(a.Calendar)

	actionn, err := a.Calendar.NextAction(a.Access)
	if err != nil {
		log.Print(err)
		return
	}

	a.candidate = actionn
}

func (a *CalendarAgent) ResponsibleActionable() (models.Actionable, error) {
	if a.candidate != nil {
		return a.Calendar.CurrentFixture(a.Access)
	} else {
		return nil, errors.New("No responsible acitonable")
	}
}

func (a *CalendarAgent) Candidate() (models.Action, bool) {
	return a.candidate, a.candidate != nil
}
