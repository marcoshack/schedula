package repository

import (
	"fmt"
	"time"

	"github.com/marcoshack/schedula/entity"
)

// JobsMySQL ...
type JobsMySQL struct {
}

// NewJobsMySQL ...
func NewJobsMySQL() (Jobs, error) {
	return nil, fmt.Errorf("repository.JobsMySQL not implemented")
}

// Add ...
func (r *JobsMySQL) Add(entity.Job) (entity.Job, error) {
	return entity.Job{}, nil
}

// Get ...
func (r *JobsMySQL) Get(id string) (entity.Job, error) {
	return entity.Job{}, nil
}

// List ...
func (r *JobsMySQL) List(skip int, limit int) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}

// Remove ...
func (r *JobsMySQL) Remove(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Cancel ...
func (r *JobsMySQL) Cancel(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// AddExecution ...
func (r *JobsMySQL) AddExecution(jobID string, date time.Time, status string, message string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Count ...
func (r *JobsMySQL) Count() int {
	return 0
}

// ListBySchedule ...
func (r *JobsMySQL) ListBySchedule(timestamp int64) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}
