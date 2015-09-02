package repository

import (
	"fmt"
	"time"

	"github.com/marcoshack/schedula/entity"
)

// JobsInMemoryWithChannels ...
type JobsInMemoryWithChannels struct {
}

// NewJobsInMemoryWithChannels ...
func NewJobsInMemoryWithChannels() (Jobs, error) {
	return nil, fmt.Errorf("repository.JobsInMemoryWithChannels not implemented")
}

// Add ...
func (r *JobsInMemoryWithChannels) Add(entity.Job) (entity.Job, error) {
	return entity.Job{}, nil
}

// Get ...
func (r *JobsInMemoryWithChannels) Get(id string) (entity.Job, error) {
	return entity.Job{}, nil
}

// List ...
func (r *JobsInMemoryWithChannels) List(skip int, limit int) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}

// Remove ...
func (r *JobsInMemoryWithChannels) Remove(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Cancel ...
func (r *JobsInMemoryWithChannels) Cancel(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// AddExecution ...
func (r *JobsInMemoryWithChannels) AddExecution(jobID string, date time.Time, status string, message string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Count ...
func (r *JobsInMemoryWithChannels) Count() int {
	return 0
}

// ListBySchedule ...
func (r *JobsInMemoryWithChannels) ListBySchedule(timestamp int64) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}
