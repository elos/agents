package agents

import (
	"github.com/elos/autonomous"
	"github.com/elos/data"
	"github.com/elos/transfer"
)

type ClientDataAgent struct {
	*autonomous.Core
	*autonomous.Identified
	data.Store

	read chan *transfer.Envelope
	transfer.SocketConnection
}

func NewClientDataAgent(c transfer.SocketConnection, s data.Store) *ClientDataAgent {
	a := &ClientDataAgent{
		Core:             autonomous.NewCore(),
		Identified:       autonomous.NewIdentified(),
		SocketConnection: c,
		Store:            s,
		read:             make(chan *transfer.Envelope),
	}

	a.SetDataOwner(c.Agent())

	return a
}

func (a *ClientDataAgent) Run() {
	a.startup()
	stopChannel := a.Core.StopChannel()
	modelsChannel := *a.Store.RegisterForUpdates(a.DataOwner())

	for {
		select {
		case e := <-a.read:
			go transfer.Route(e, a.Store)
		case p := <-modelsChannel:
			a.WriteJSON(p)
		case _ = <-*stopChannel:
			a.shutdown()
			break
		}
	}
}

func (a *ClientDataAgent) startup() {
	a.Core.Startup()
	go ReadSocketConnection(a.SocketConnection, &a.read, a.Core.StopChannel())
}

func (a *ClientDataAgent) shutdown() {
	a.Core.Shutdown()
}

func ReadSocketConnection(c transfer.SocketConnection, rc *chan *transfer.Envelope, endChannel *chan bool) {
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

	*endChannel <- true
}

func (a *ClientDataAgent) WriteJSON(v interface{}) {
	a.SocketConnection.WriteJSON(v)
}
