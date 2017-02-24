package repository

import (
	"fmt"
	"time"

	"github.com/marcoshack/schedula/entity"
)

// JobsRedis ...
type JobsRedis struct {
}

// NewJobsRedis ...
func NewJobsRedis() (Jobs, error) {
	return nil, fmt.Errorf("repository.JobsRedis not implemented")
}

// Add ...
func (r *JobsRedis) Add(entity.Job) (entity.Job, error) {
	return entity.Job{}, nil
}

// Get ...
func (r *JobsRedis) Get(id string) (entity.Job, error) {
	return entity.Job{}, nil
}

// List ...
func (r *JobsRedis) List(skip int, limit int) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}

// Remove ...
func (r *JobsRedis) Remove(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Cancel ...
func (r *JobsRedis) Cancel(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// AddExecution ...
func (r *JobsRedis) AddExecution(jobID string, date time.Time, status string, message string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Count ...
func (r *JobsRedis) Count() int {
	return 0
}

// ListBySchedule ...
func (r *JobsRedis) ListBySchedule(timestamp int64) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}
