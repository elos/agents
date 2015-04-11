package agents

import (
	"fmt"
	"io"
	"log"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/models"
	"github.com/elos/models/interactive"
	"github.com/elos/models/persistence"
	"github.com/elos/transfer"
	"github.com/robertkrimen/otto"
)

type REPLAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper
	data.Access

	transfer.SocketConnection

	space *interactive.Space
	otto  *otto.Otto
}

func NewREPLAgent(c transfer.SocketConnection, access data.Access) *REPLAgent {
	a := new(REPLAgent)

	a.Life = autonomous.NewLife()
	a.Stopper = make(autonomous.Stopper)
	a.SocketConnection = c
	a.Access = access

	s, _ := interactive.Store(persistence.ModelsStore(access), access.Client().(models.User))

	a.space = s
	a.otto = otto.New()

	a.space.Expose(a.otto)

	return a
}

func (a *REPLAgent) Start() {
	a.SocketConnection.WriteJSON(&output{fmt.Sprintf("Welcome to Elos, %s", a.space.User.Name)})
	go a.read(a.SocketConnection, a.Stopper)

	a.Life.Begin()
	<-a.Stopper
	a.SocketConnection.Close()
	a.Life.End()
}

type command struct {
	Command string `json:"command"`
}

type output struct {
	Output string `json:"output"`
}

func (a *REPLAgent) read(c transfer.SocketConnection, closed chan bool) {
	log.Print("reading")

Read:
	for {
		c := new(command)
		err := a.SocketConnection.ReadJSON(c)
		if err != nil {
			if err != io.EOF {
				log.Printf("An error occurred while reading a socket, err: %s", err)
			}
			break Read
		}

		if len(c.Command) > 0 {
			a.interpret(c.Command)
		}
	}

	closed <- true
}

func (a *REPLAgent) interpret(entry string) {
	if len(entry) == 0 {
		return
	}
	value, err := a.otto.Run(entry)

	var s string
	if err != nil {
		s = err.Error()
	} else {
		s = fmt.Sprintf("%v", value)
	}

	if len(s) > 0 {
		fmt.Println(s)
	}

	a.SocketConnection.WriteJSON(&output{s})
}
