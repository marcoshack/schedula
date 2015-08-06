package schedula

import "fmt"

// Scheduler is a service to schedule jobs
type Scheduler interface {
	Schedule(job *Job) (id string, err error)
	Type() string
}

// InitScheduler creates a new Scheduler instance of the given type.
// Currently acceptable values for 'schedulerType' are: "in-memory"
func InitScheduler(schedulerType string) (Scheduler, error) {
	if schedulerType == "in-memory" {
		return &InMemoryScheduler{}, nil
	}
	return nil, fmt.Errorf("schedula: invalid scheduler type during initialization: '%s'", schedulerType)
}

// InMemoryScheduler implements Scheduler interface using non-replicated in-memory data structure.
// This is a example implementation and should be used only for test purposes.
type InMemoryScheduler struct {
}

// Schedule ...
func (s *InMemoryScheduler) Schedule(job *Job) (string, error) {
	return "fake-job-1", nil
}

// Type ...
func (s *InMemoryScheduler) Type() string {
	return "in-memory"
}
