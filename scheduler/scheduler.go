package scheduler

import (
	"fmt"
	"time"

	"github.com/marcoshack/schedula/callback"
	"github.com/marcoshack/schedula/entity"
	"github.com/marcoshack/schedula/repository"
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

// Config holds Scheduler configuration parameters
type Config struct {
	NumberOfWorkers int
}

// New creates a Scheduler instance of the given type.
// Currently acceptable values for 'schedulerType' are: "in-memory"
func New(t string, r repository.Repository, e callback.Executor, c Config) (Scheduler, error) {
	switch t {
	case "ticker":
		return &TickerScheduler{
			Config:           c,
			jobs:             r,
			callbackExecutor: e,
			tickInterval:     DefaultTickInterval * time.Second,
			callbackChannel:  make(chan entity.Job, 10000),
		}, nil
	}
	return nil, fmt.Errorf("invalid scheduler type: '%s'", t)
}

// StartNew creates and starts a Scheduler instance
func StartNew(t string, r repository.Repository, e callback.Executor, c Config) (Scheduler, error) {
	scheduler, initErr := New(t, r, e, c)
	if initErr != nil {
		return nil, initErr
	}

	startErr := scheduler.Start()
	if startErr != nil {
		return nil, startErr
	}
	return scheduler, nil
}