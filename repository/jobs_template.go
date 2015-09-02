package repository

import (
	"fmt"
	"time"

	"github.com/marcoshack/schedula/entity"
)

// JobsTemplate ...
type JobsTemplate struct {
}

// NewJobsTemplate ...
func NewJobsTemplate() (Jobs, error) {
	return nil, fmt.Errorf("repository.JobsTemplate not implemented")
}

// Add ...
func (r *JobsTemplate) Add(entity.Job) (entity.Job, error) {
	return entity.Job{}, nil
}

// Get ...
func (r *JobsTemplate) Get(id string) (entity.Job, error) {
	return entity.Job{}, nil
}

// List ...
func (r *JobsTemplate) List(skip int, limit int) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}

// Remove ...
func (r *JobsTemplate) Remove(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Cancel ...
func (r *JobsTemplate) Cancel(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// AddExecution ...
func (r *JobsTemplate) AddExecution(jobID string, date time.Time, status string, message string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Count ...
func (r *JobsTemplate) Count() int {
	return 0
}

// ListBySchedule ...
func (r *JobsTemplate) ListBySchedule(timestamp int64) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}
