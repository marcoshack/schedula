package schedula

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

const (
	// TypeInMemory holds the value for scheduler type 'in-memory'
	TypeInMemory = "in-memory"

	// DefaultTickInterval is the number of seconds for scheduler's ticker
	DefaultTickInterval = 1
)

// Scheduler is a service to schedule jobs
type Scheduler interface {
	Add(Job) (Job, error)
	Get(id string) (Job, error)
	List(size int, skip int) ([]Job, error)
	Size() int
	Type() string
	Start() error
	Stop() error
}

// InMemoryScheduler implements Scheduler interface using non-replicated in-memory data structure.
// This is a example implementation and should be used only for test purposes.
type InMemoryScheduler struct {
	TickInterval time.Duration
	ticker       *time.Ticker
	httpClient   *http.Client
	jobs         JobMap
	params       map[string]interface{}
}

// JobMap ...
type JobMap struct {
	sync.RWMutex
	id    map[string]*Job
	sched map[int64][]*Job
	list  []Job
}

// InitScheduler creates a new Scheduler instance of the given type.
// Currently acceptable values for 'schedulerType' are: "in-memory"
func InitScheduler(schedulerType string, params map[string]interface{}) (Scheduler, error) {
	switch schedulerType {
	case TypeInMemory:
		return &InMemoryScheduler{
			TickInterval: DefaultTickInterval * time.Second,
			params:       initParams(params),
		}, nil
	}
	return nil, fmt.Errorf("scheduler: invalid scheduler type: '%s'", schedulerType)
}

func initParams(params map[string]interface{}) map[string]interface{} {
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

// StartScheduler ...
func StartScheduler(schedulerType string, params map[string]interface{}) (Scheduler, error) {
	scheduler, initErr := InitScheduler(schedulerType, params)
	if initErr != nil {
		return nil, initErr
	}

	startErr := scheduler.Start()
	if startErr != nil {
		return nil, startErr
	}
	return scheduler, nil
}

// Add ...
func (s *InMemoryScheduler) Add(job Job) (Job, error) {
	job.ID = uuid.New()
	s.jobs.Lock()
	defer s.jobs.Unlock()

	// add to job list
	s.jobs.list = append(s.jobs.list, job)

	// add to jobs map by ID
	s.jobs.id[job.ID] = &job

	// add to jobs map by schedule timetamp
	jobTimestamp, err := job.Schedule.NextTimestamp()
	if err != nil {
		return Job{}, err
	}
	if s.jobs.sched[jobTimestamp] == nil {
		s.jobs.sched[jobTimestamp] = make([]*Job, 0)
	}
	s.jobs.sched[jobTimestamp] = append(s.jobs.sched[jobTimestamp], &job)

	return job, nil
}

// Get returns the Job associated with the given id or nil if it doensn't exist
func (s *InMemoryScheduler) Get(id string) (Job, error) {
	s.jobs.RLock()
	defer s.jobs.RUnlock()
	if s.jobs.id[id] != nil {
		return *s.jobs.id[id], nil
	}
	return Job{}, nil
}

// List returns the list of scheduled jobs
func (s *InMemoryScheduler) List(skip int, limit int) ([]Job, error) {
	s.jobs.RLock()
	defer s.jobs.RUnlock()

	if len(s.jobs.id) == 0 || skip > len(s.jobs.id) || limit < 0 {
		return make([]Job, 0), nil
	}

	start := skip
	if start > len(s.jobs.list) {
		skip = len(s.jobs.list)
	}
	end := skip + limit
	if end > len(s.jobs.list) {
		end = len(s.jobs.list)
	}

	return s.jobs.list[start:end], nil
}

// Size returns the number of active jobs in the scheduler
func (s *InMemoryScheduler) Size() int {
	s.jobs.RLock()
	defer s.jobs.RUnlock()
	return len(s.jobs.list)
}

// Type ...
func (s *InMemoryScheduler) Type() string {
	return "in-memory"
}

// Start ...
func (s *InMemoryScheduler) Start() error {
	if s.ticker != nil {
		return fmt.Errorf("scheduler: scheduler already started")
	}
	s.ticker = time.NewTicker(s.TickInterval)
	s.httpClient = &http.Client{}
	s.jobs = JobMap{
		id:    make(map[string]*Job),
		sched: make(map[int64][]*Job),
	}

	if s.params["no-tick-log"].(bool) == false {
		log.Printf("scheduler: start ticking every %d seconds", s.TickInterval/time.Second)
	}
	go s.tick()
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

func (s *InMemoryScheduler) tick() {
	for now := range s.ticker.C {
		s.jobs.RLock()
		schedList := s.jobs.sched[now.Unix()]
		s.jobs.RUnlock()
		if schedList != nil {
			log.Printf("scheduler: executing %d callbacks at %v (%v)", len(schedList), now.Unix(), now)
			for _, job := range schedList {
				go s.execute(job)
			}
		}
	}
}

func (s *InMemoryScheduler) execute(job *Job) {
	var body = new(bytes.Buffer)
	encErr := json.NewEncoder(body).Encode(job)
	if encErr != nil {
		log.Printf("scheduler: unable to encode request body for job %s: %v", job.ID, encErr)
	}

	req, reqErr := http.NewRequest("POST", job.CallbackURL, body)
	req.Header.Set("User-Agent", "schedula")
	if reqErr != nil {
		log.Printf("scheduler: error creating callback request for job %s: %v", job.ID, reqErr)
		return
	}

	res, postErr := s.httpClient.Do(req)
	if postErr != nil {
		log.Printf("scheduler: error on job %s callback: %v", job.ID, postErr)
		return
	}

	if res.StatusCode == http.StatusOK {
		log.Printf("schduler: job %s callback succeed: %s", job.ID, res.Status)
	}
}
