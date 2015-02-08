package agents

import (
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/data"
)

var DefaultSleepAgentStartPeriod time.Duration = 10 * time.Second

type SleepAgent struct {
	autonomous.Life
	autonomous.Managed
	autonomous.Tallied
	autonomous.Stopper

	startPeriod time.Duration
	ticker      *time.Ticker

	DB data.DB
}

func NewSleepAgent(db data.DB, a data.Identifiable, d time.Duration) autonomous.Agent {
	s := new(SleepAgent)

	s.Life = autonomous.NewLife()
	s.startPeriod = d
	s.DB = db

	return s
}

func (s *SleepAgent) Start() {
	s.ticker = time.NewTicker(s.startPeriod)
	go s.Go()

	s.Life.Begin()

Run:
	for {
		select {
		case <-s.ticker.C:
			go s.Go()
		case <-s.Stopper:
			break Run
		}
	}

	s.ticker.Stop()
	s.ticker = nil

	s.Life.End()
}

func (s *SleepAgent) Go() {
	s.Tallied.Incr()
	// implement what the sleep agent would actually do
	s.Tallied.Decr()
}
