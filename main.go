package agents

import (
	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
)

type AgentConstructor func(data.DB, *models.User) autonomous.Agent

var AgentOptions = map[string]AgentConstructor{
	"action": func(db data.DB, u *models.User) autonomous.Agent {
		return NewActionAgent(db, u)
	},
}

type MainAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	autonomous.Manager

	data.DB
	*models.User
}

func NewMainAgent(db data.DB, u *models.User) *MainAgent {
	h := autonomous.NewHub()
	go h.Start()

	return &MainAgent{
		Life:    autonomous.NewLife(),
		Stopper: make(autonomous.Stopper),
		Manager: h,

		DB:   db,
		User: u,
	}
}

func (a *MainAgent) Start() {
	a.Life.Begin()

	for _, constructor := range AgentOptions {
		go a.Manager.StartAgent(constructor(a.DB, a.User))
	}

	<-a.Stopper
	a.shutdown()
}

func (a *MainAgent) shutdown() {
	go a.Manager.Stop()
	a.Manager.WaitStop()
	a.Life.End()
}
