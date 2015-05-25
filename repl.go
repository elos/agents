package agents

import (
	"fmt"
	"io"
	"log"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/ehttp/sock"
	"github.com/elos/interactive"
	"github.com/elos/models"
)

type REPLAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	data.DB
	sock.Conn
	env *interactive.Env
	*models.User
}

func NewREPLAgent(c sock.Conn, db data.DB, u *models.User) *REPLAgent {
	a := new(REPLAgent)

	a.Life = autonomous.NewLife()
	a.Stopper = make(autonomous.Stopper)

	a.Conn = c
	a.DB = db
	a.env = interactive.NewEnv(db, u)
	a.User = u

	return a
}

func (a *REPLAgent) Start() {
	a.Conn.WriteJSON(&output{fmt.Sprintf("Welcome to Elos Interactive REPL, %s", a.User.Name)})
	a.Conn.WriteJSON(&output{fmt.Sprintf("Try typing \"me\"")})

	go a.read(a.Conn, a.Stopper)
	a.Life.Begin()
	<-a.Stopper
	// a.Conn.Close() yikes, might need to add to the interface
	a.Life.End()
}

type command struct {
	Command string `json:"command"`
}

type output struct {
	Output string `json:"output"`
}

func (a *REPLAgent) read(c sock.Conn, closed chan bool) {
Read:
	for {
		c := new(command)
		err := a.Conn.ReadJSON(c)
		if err != nil {
			if err != io.EOF {
				log.Printf("An error occurred while reading a socket, err: %s", err)
			}
			break Read
		}

		if len(c.Command) > 0 {
			a.Conn.WriteJSON(&output{a.env.Interpret(c.Command)})
		}
	}

	closed <- true
}
