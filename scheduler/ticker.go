package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/marcoshack/schedula/callback"
	"github.com/marcoshack/schedula/entity"
	"github.com/marcoshack/schedula/repository"
)

// TickerScheduler implements Scheduler interface using non-replicated in-memory data structure.
// This is a example implementation and should be used only for test purposes.
type TickerScheduler struct {
	Config           Config
	tickInterval     time.Duration
	ticker           *time.Ticker
	jobs             repository.Repository
	callbackExecutor callback.Executor
	callbackChannel  chan entity.Job
}

// Start ...
func (s *TickerScheduler) Start() error {
	if s.ticker != nil {
		return fmt.Errorf("scheduler: scheduler already started")
	}
	s.ticker = time.NewTicker(s.tickInterval)

	for i := 0; i < s.Config.NumberOfWorkers; i++ {
		go s.workerLoop(s.callbackChannel)
	}

	go s.tickerLoop()
	return nil
}

// Stop ...
func (s *TickerScheduler) Stop() error {
	if s.ticker == nil {
		return fmt.Errorf("scheduler: scheduler wasn't started, cannot stop")
	}
	s.ticker.Stop()
	return nil
}

func (s *TickerScheduler) tickerLoop() {
	for now := range s.ticker.C {
		go s.publishJobs(now)
	}
}

func (s *TickerScheduler) publishJobs(now time.Time) {
	jobs, err := s.jobs.ListBySchedule(now.Unix())
	if err != nil {
		log.Printf("scheduler: error retrieving job list scheduled at %v: %v", now, err)
		return
	}

	if jobs == nil || len(jobs) == 0 {
		return
	}

	log.Printf("scheduler: launching %d callbacks scheduled at %v (%v)", len(jobs), now.Unix(), now)
	for _, job := range jobs {
		if job.IsExecutable() {
			s.callbackChannel <- job
		}
	}
}

func (s *TickerScheduler) workerLoop(jobs chan entity.Job) {
	for job := range jobs {
		var newStatus string
		var errMessage string
		if err := s.callbackExecutor.Execute(job); err == nil {
			newStatus = entity.JobStatusSuccess
		} else {
			newStatus = entity.JobStatusError
			errMessage = fmt.Sprintf("%v", err)
			log.Printf("scheduler: job[ID:%s]: error executing callback: %v", job.ID, err)
		}
		if _, err := s.jobs.AddExecution(job.ID, time.Now(), newStatus, errMessage); err != nil {
			log.Printf("scheduler: job[ID:%s]: error adding execution with status '%s': %v", job.ID, newStatus, err)
		}
	}
}
