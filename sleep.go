package agents

import (
	"log"
	"time"

	"github.com/elos/autonomous"
	"github.com/elos/data"
)

var DefaultSleepAgentStartPeriod time.Duration = 10 * time.Second

type SleepAgent struct {
	*autonomous.Core

	startPeriod time.Duration
	ticker      *time.Ticker

	DB data.DB
}

func NewSleepAgent(db data.DB, a data.Identifiable, d time.Duration) autonomous.Agent {
	return &SleepAgent{
		Core:        autonomous.NewCore(),
		startPeriod: d,
		DB:          db,
	}
}

func (s *SleepAgent) Run() {
	s.startup()
	stopChannel := s.Core.StopChannel()

	for {
		select {
		case _ = <-s.ticker.C:
			go s.Go()
		case _ = <-*stopChannel:
			s.shutdown()
			break
		}
	}
}

func (s *SleepAgent) startup() {
	s.Core.Startup()
	s.ticker = time.NewTicker(s.startPeriod)
	go s.Go()
}

func (s *SleepAgent) shutdown() {
	s.ticker.Stop()
	s.ticker = nil
	s.Core.Shutdown()
}

func (s *SleepAgent) Go() {
	s.IncrementProcesses()
	log.Print("This is the sleep agent checking in")
	// implement what the sleep agent would actually do
	s.DecrementProcesses()
}
