package agents

import (
	"log"

	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/transfer"
)

type ClientDataAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper
	*data.Access

	read chan *transfer.Envelope
	transfer.SocketConnection
}

func NewClientDataAgent(c transfer.SocketConnection, access *data.Access) *ClientDataAgent {
	a := new(ClientDataAgent)

	a.Life = autonomous.NewLife()
	a.SocketConnection = c
	a.Access = access

	return a
}

func (a *ClientDataAgent) Start() {
	var mc chan *data.Change = *a.RegisterForUpdates(a.Client)
	go ReadSocketConnection(a.SocketConnection, a.read, a.Stopper)
	a.Life.Begin()

Run:
	for {
		select {
		case e := <-a.read:
			go transfer.Route(e, a.Access)
		case c := <-mc:
			a.WriteJSON(c)
		case <-a.Stopper:
			break Run
		}
	}

	a.Life.End()
}

func ReadSocketConnection(c transfer.SocketConnection, rc chan *transfer.Envelope, closed chan bool) {
	// TODO add read limit and deadline

Read:
	for {
		e := new(transfer.Envelope)
		e.Connection = c

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
