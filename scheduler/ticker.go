package scheduler

import (
	"fmt"
	"log"
	"net/url"
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
	jobs             repository.Jobs
	callbackExecutor callback.Executor
	HostContexts     map[string]*HostContext
}

// HostContext ...
type HostContext struct {
	Host     string
	C        chan entity.Job
	LastUsed time.Time
	Workers  int
}

// NewTickerScheduler ...
func NewTickerScheduler(r repository.Jobs, e callback.Executor, c Config) *TickerScheduler {
	return &TickerScheduler{
		Config:           c,
		jobs:             r,
		callbackExecutor: e,
		tickInterval:     DefaultTickInterval * time.Second,
		HostContexts:     make(map[string]*HostContext),
	}
}

// Start ...
func (s *TickerScheduler) Start() error {
	if s.ticker != nil {
		return fmt.Errorf("scheduler: scheduler already started")
	}
	s.ticker = time.NewTicker(s.tickInterval)
	go s.tick()
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

func (s *TickerScheduler) tick() {
	for now := range s.ticker.C {
		jobs, err := s.jobs.ListBySchedule(now.Unix())
		if err != nil {
			log.Printf("scheduler: error retrieving job list scheduled at %v: %v", now, err)
			continue
		}

		if len(jobs) > 0 {
			log.Printf("scheduler: launching %d callbacks scheduled at %v (%v)", len(jobs), now.Unix(), now)
			go s.publish(jobs)
		}
	}
}

func (s *TickerScheduler) publish(jobs []entity.Job) {
	for _, job := range jobs {
		if !job.IsExecutable() {
			continue
		}

		context, err := s.context(job)
		if err != nil {
			log.Printf("scheduler: error publishing job callback: %v", err)
		}
		context.C <- job
	}
}

func (s *TickerScheduler) context(job entity.Job) (*HostContext, error) {
	url, err := url.ParseRequestURI(job.CallbackURL)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve host context, error parsing callback URL: %v", err)
	}

	var context *HostContext
	if _, exists := s.HostContexts[url.Host]; !exists {
		context = &HostContext{
			Host:     url.Host,
			C:        make(chan entity.Job, 1000), // TODO think better about the host channel buffer size
			LastUsed: time.Now(),
			Workers:  s.Config.WorkersPerHost,
		}
		s.HostContexts[url.Host] = context
		for i := 0; i < context.Workers; i++ {
			go s.handle(context)
		}
		log.Printf("scheduler: host context for %s created with %d workers", context.Host, context.Workers)
	}
	return context, nil
}

func (s *TickerScheduler) handle(context *HostContext) {
	for job := range context.C {
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
