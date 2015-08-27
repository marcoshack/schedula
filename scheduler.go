package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	// DefaultTickInterval is the number of seconds for scheduler's ticker
	DefaultTickInterval = 1

	// DefaultNumberOfWorkers is the number of go routines spawned to execute callbacks for each tick
	DefaultNumberOfWorkers = 50
)

// Scheduler is a service to schedule jobs
type Scheduler interface {
	Start() error
	Stop() error
}

// SchedulerConfig holds Scheduler configuration parameters
type SchedulerConfig struct {
	NumberOfWorkers int
}

// TickerScheduler implements Scheduler interface using non-replicated in-memory data structure.
// This is a example implementation and should be used only for test purposes.
type TickerScheduler struct {
	Config SchedulerConfig

	tickInterval    time.Duration
	ticker          *time.Ticker
	httpClient      *http.Client
	jobs            Repository
	callbackChannel chan *Job
}

// InitScheduler creates a new Scheduler instance of the given type.
// Currently acceptable values for 'schedulerType' are: "in-memory"
func InitScheduler(repo Repository, config SchedulerConfig) (Scheduler, error) {
	return &TickerScheduler{
		Config:          config,
		jobs:            repo,
		httpClient:      &http.Client{},
		tickInterval:    DefaultTickInterval * time.Second,
		callbackChannel: make(chan *Job, 10000),
	}, nil
}

// InitAndStartScheduler ...
func InitAndStartScheduler(repo Repository, config SchedulerConfig) (Scheduler, error) {
	scheduler, initErr := InitScheduler(repo, config)
	if initErr != nil {
		return nil, initErr
	}

	startErr := scheduler.Start()
	if startErr != nil {
		return nil, startErr
	}
	return scheduler, nil
}

// Start ...
func (s *TickerScheduler) Start() error {
	if s.ticker != nil {
		return fmt.Errorf("scheduler: scheduler already started")
	}
	s.ticker = time.NewTicker(s.tickInterval)

	for i := 0; i < s.Config.NumberOfWorkers; i++ {
		go s.executeJobs(s.callbackChannel)
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

	if jobs != nil && len(jobs) != 0 {
		log.Printf("scheduler: launching %d callbacks scheduled at %v (%v)", len(jobs), now.Unix(), now)
		for _, job := range jobs {
			if job.IsExecutable() {
				s.callbackChannel <- job
			}
		}
	}
}

func (s *TickerScheduler) executeJobs(jobs chan *Job) {
	for job := range jobs {
		s.execute(job)
	}
}

func (s *TickerScheduler) execute(job *Job) {
	req, reqErr := s.createCallbackRequest(job.CallbackURL, job)
	if reqErr != nil {
		log.Printf("scheduler: job[ID:%s]: error creating callback request: %v", job.ID, reqErr)
		job.Status = JobStatusError
		return
	}

	res, postErr := s.httpClient.Do(req)
	if postErr != nil {
		log.Printf("scheduler: job[ID:%s]: callback error: %v", job.ID, postErr)
		job.Status = JobStatusError
		return
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		log.Printf("scheduler: job[ID:%s]: callback failed: %s", job.ID, res.Status)
		job.Status = JobStatusFail
		return
	}

	job.Status = JobStatusSuccess
}

func (s *TickerScheduler) createCallbackRequest(urlStr string, job *Job) (*http.Request, error) {
	var body = new(bytes.Buffer)
	encErr := json.NewEncoder(body).Encode(job)
	if encErr != nil {
		return nil, fmt.Errorf("unable to encode request body: %v", encErr)
	}

	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "schedula")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
