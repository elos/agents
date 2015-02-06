package agents

import (
	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/transfer"
)

type ClientDataAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Stopper
	*autonomous.Identified
	data.Store

	read chan *transfer.Envelope
	transfer.SocketConnection
}

func NewClientDataAgent(c transfer.SocketConnection, s data.Store) *ClientDataAgent {
	a := new(ClientDataAgent)
	a.Identified = autonomous.NewIdentified()
	a.Life = autonomous.NewLife()
	a.SocketConnection = c
	a.Store = s
	a.SetDataOwner(c.Agent())

	return a
}

func (a *ClientDataAgent) Run() {
	modelsChannel := *a.Store.RegisterForUpdates(a.DataOwner())
	a.startup()
	a.Life.Begin()
Run:
	for {
		select {
		case e := <-a.read:
			go transfer.Route(e, a.Store)
		case p := <-modelsChannel:
			a.WriteJSON(p)
		case <-a.Stopper:
			break Run
		}
	}

	a.shutdown()
	a.Life.End()
}

func (a *ClientDataAgent) startup() {
	go ReadSocketConnection(a.SocketConnection, &a.read, a.Stopper)
}

func (a *ClientDataAgent) shutdown() {
}

func ReadSocketConnection(c transfer.SocketConnection, rc *chan *transfer.Envelope, closed chan bool) {
	// TODO add read limit and deadline
	for {
		var e transfer.Envelope

		err := c.ReadJSON(&e)

		if err != nil {
			//Logf("An error occurred while reading a transferection, err: %s", err)

			/*
				If there was an error break inf. loop.
				Function then completes, and endChannel is called
			*/
			break
		}

		e.Connection = c

		*rc <- &e
	}

	closed <- true
}

func (a *ClientDataAgent) WriteJSON(v interface{}) {
	a.SocketConnection.WriteJSON(v)
}
