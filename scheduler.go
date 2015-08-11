package schedula

import (
	"fmt"
	"log"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

const (
	// TypeInMemory holds the value for scheduler type 'in-memory'
	TypeInMemory = "in-memory"
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
	jobs         JobMap
	params       map[string]interface{}
}

// JobMap ...
type JobMap struct {
	sync.RWMutex
	m map[string]*Job
	l []Job
}

// InitScheduler creates a new Scheduler instance of the given type.
// Currently acceptable values for 'schedulerType' are: "in-memory"
func InitScheduler(schedulerType string, params map[string]interface{}) (Scheduler, error) {
	switch schedulerType {
	case TypeInMemory:
		return &InMemoryScheduler{
			TickInterval: 10 * time.Second,
			params:       initParams(params),
		}, nil
	}
	return nil, fmt.Errorf("schedula: invalid scheduler type: '%s'", schedulerType)
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
	s.jobs.l = append(s.jobs.l, job)
	s.jobs.m[job.ID] = &job
	return job, nil
}

// Get returns the Job associated with the given id or nil if it doensn't exist
func (s *InMemoryScheduler) Get(id string) (Job, error) {
	s.jobs.RLock()
	defer s.jobs.RUnlock()
	if s.jobs.m[id] != nil {
		return *s.jobs.m[id], nil
	}
	return Job{}, nil
}

// List returns the list of scheduled jobs
func (s *InMemoryScheduler) List(skip int, limit int) ([]Job, error) {
	s.jobs.RLock()
	defer s.jobs.RUnlock()
	var end = limit

	if len(s.jobs.m) == 0 || skip > len(s.jobs.m) || limit < 0 {
		return make([]Job, 0), nil
	}

	if limit > len(s.jobs.m) {
		end = len(s.jobs.m)
	}

	return s.jobs.l[skip:end], nil
}

// Size returns the number of active jobs in the scheduler
func (s *InMemoryScheduler) Size() int {
	s.jobs.RLock()
	defer s.jobs.RUnlock()
	return len(s.jobs.l)
}

// Type ...
func (s *InMemoryScheduler) Type() string {
	return "in-memory"
}

// Start ...
func (s *InMemoryScheduler) Start() error {
	if s.ticker != nil {
		return fmt.Errorf("schedula: scheduler already started")
	}
	s.ticker = time.NewTicker(s.TickInterval)
	s.jobs = JobMap{m: make(map[string]*Job)}

	if s.params["no-tick-log"].(bool) == false {
		log.Printf("InMemoryScheduler: start ticking every %d seconds", s.TickInterval/time.Second)
	}
	go s.tick()
	return nil
}

// Stop ...
func (s *InMemoryScheduler) Stop() error {
	if s.ticker == nil {
		return fmt.Errorf("schedula: scheduler wasn't started, cannot stop")
	}
	s.ticker.Stop()
	return nil
}

func (s *InMemoryScheduler) tick() {
	for now := range s.ticker.C {
		if s.params["no-tick-log"].(bool) == false {
			log.Printf("InMemoryScheduler: tick %s", now)
			// TODO execute jobs callbacks
		}
	}
}
