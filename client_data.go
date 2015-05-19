package agents

import (
	"log"

	"github.com/elos/api/hermes"
	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/data/transfer"
	"github.com/elos/ehttp/sock"
)

type ClientDataAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper

	data.DB
	read chan *transfer.Envelope
	sock.Conn
}

func NewClientDataAgent(c sock.Conn, db data.DB) *ClientDataAgent {
	a := new(ClientDataAgent)

	a.Life = autonomous.NewLife()
	a.Conn = c
	a.DB = db
	a.read = make(chan *transfer.Envelope)

	return a
}

func (a *ClientDataAgent) Start() {
	changes := data.Filter(a.DB.Changes(), func(change *data.Change) bool {
		return true
	})

	go ReadSocketConnection(a.Conn, a.read, a.Stopper)

	a.Life.Begin()

Run:
	for {
		select {
		case e := <-a.read:
			go hermes.Serve(e, a.DB)
		case change := <-*changes:
			a.Conn.WriteJSON(change)
		case <-a.Stopper:
			break Run
		}
	}

	a.Life.End()
}

func ReadSocketConnection(c sock.Conn, rc chan *transfer.Envelope, closed chan bool) {
	// TODO add read limit and deadline

Read:
	for {
		e := new(transfer.Envelope)
		e.Conn = c

		err := c.ReadJSON(e)

		if err != nil {
			log.Printf("An error occurred while reading a socket, err: %s", err)

			// If there was an error break inf. loop.
			// function then completes
			break Read
		}

		rc <- e
	}

	closed <- true
}
