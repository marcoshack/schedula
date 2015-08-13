package schedula

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	// DefaultTickInterval is the number of seconds for scheduler's ticker
	DefaultTickInterval = 1
)

// Scheduler is a service to schedule jobs
type Scheduler interface {
	Start() error
	Stop() error
	Add(Job) (Job, error)
	Get(id string) (Job, error)
	List(skip int, limit int) ([]Job, error)
	Count() int
}

// InMemoryScheduler implements Scheduler interface using non-replicated in-memory data structure.
// This is a example implementation and should be used only for test purposes.
type InMemoryScheduler struct {
	tickInterval time.Duration
	ticker       *time.Ticker
	httpClient   *http.Client
	jobs         Repository
	params       map[string]interface{}
}

// InitScheduler creates a new Scheduler instance of the given type.
// Currently acceptable values for 'schedulerType' are: "in-memory"
func InitScheduler(repo Repository, params map[string]interface{}) (Scheduler, error) {
	return &InMemoryScheduler{
		tickInterval: DefaultTickInterval * time.Second,
		jobs:         repo,
		params:       parseParams(params),
	}, nil
}

// InitAndStartScheduler ...
func InitAndStartScheduler(repo Repository, params map[string]interface{}) (Scheduler, error) {
	scheduler, initErr := InitScheduler(repo, params)
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
func (s *InMemoryScheduler) Start() error {
	if s.ticker != nil {
		return fmt.Errorf("scheduler: scheduler already started")
	}
	s.ticker = time.NewTicker(s.tickInterval)
	s.httpClient = &http.Client{}

	if s.params["no-tick-log"].(bool) == false {
		log.Printf("scheduler: start ticking every %d seconds", s.tickInterval/time.Second)
	}
	go s.tickerLoop()
	return nil
}

// Stop ...
func (s *InMemoryScheduler) Stop() error {
	if s.ticker == nil {
		return fmt.Errorf("scheduler: scheduler wasn't started, cannot stop")
	}
	s.ticker.Stop()
	return nil
}

// Add ...
func (s *InMemoryScheduler) Add(job Job) (Job, error) {
	return s.jobs.Add(job)
}

// Get ...
func (s *InMemoryScheduler) Get(id string) (Job, error) {
	return s.jobs.Get(id)
}

// List ...
func (s *InMemoryScheduler) List(skip int, limit int) ([]Job, error) {
	return s.jobs.List(skip, limit)
}

// Count ...
func (s *InMemoryScheduler) Count() int {
	return s.jobs.Count()
}

func parseParams(params map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	defaults := make(map[string]interface{})
	defaults["no-tick-log"] = false
	for key, defaultValue := range defaults {
		if params[key] != nil {
			res[key] = params[key]
		} else {
			params[key] = defaultValue
		}
	}
	return params
}

func (s *InMemoryScheduler) tickerLoop() {
	for now := range s.ticker.C {
		go s.executeCallbacks(now)
	}
}

func (s *InMemoryScheduler) executeCallbacks(now time.Time) {
	schedList, err := s.jobs.ListBySchedule(now.Unix())
	if err != nil {
		log.Printf("scheduler: error retrieving job list scheduled at %v: %v", now, err)
		return
	}

	if schedList != nil {
		log.Printf("scheduler: executing %d callbacks scheduled at %v (%v)", len(schedList), now.Unix(), now)
		for _, job := range schedList {
			go s.executeCallback(job)
		}
	}
}

func (s *InMemoryScheduler) executeCallback(job *Job) {
	var body = new(bytes.Buffer)
	encErr := json.NewEncoder(body).Encode(job)
	if encErr != nil {
		log.Printf("scheduler: unable to encode request body for job[ID:%s]: %v", job.ID, encErr)
		job.Status = JobStatusError
	}

	req, reqErr := s.createCallbackRequest(job.CallbackURL, body)
	if reqErr != nil {
		log.Printf("scheduler: error creating callback request for job[ID:%s]: %v", job.ID, reqErr)
		job.Status = JobStatusError
		return
	}

	res, postErr := s.httpClient.Do(req)
	if postErr != nil {
		log.Printf("scheduler: callback error on job[ID:%s]: %v", job.ID, postErr)
		job.Status = JobStatusError
		return
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		log.Printf("schduler: job[ID:%s] callback failed: %s", job.ID, res.Status)
		job.Status = JobStatusFail
		return
	}

	log.Printf("schduler: job[ID:%s] callback succeed: %s", job.ID, res.Status)
	job.Status = JobStatusSuccess
}

func (s *InMemoryScheduler) createCallbackRequest(urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "schedula")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
