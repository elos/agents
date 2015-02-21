package agents

import (
	"sync"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
)

type AgentConstructor func(*data.Access, models.User) autonomous.Agent

var AgentOptions = map[string]AgentConstructor{
	"action": func(a *data.Access, u models.User) autonomous.Agent {
		return NewActionAgent(a, u)
	},
}

type MainAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper
	*autonomous.Hub

	*data.Access
	models.User
}

func NewMainAgent(a *data.Access, u models.User) *MainAgent {
	return &MainAgent{
		Access:  a,
		User:    u,
		Life:    autonomous.NewLife(),
		Stopper: make(autonomous.Stopper),
		Hub:     autonomous.NewHub(),
	}
}

func (a *MainAgent) Start() {
	a.Life.Begin()
	go a.Hub.Start()

	for _, constructor := range AgentOptions {
		go a.Hub.StartAgent(constructor(a.Access, a.User))
	}

	<-a.Stopper
	a.shutdown()
}

func (a *MainAgent) shutdown() {
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		a.Hub.WaitStop()
		wg.Done()
	}()

	a.Hub.Stop()

	wg.Wait()
	a.Life.End()
}
