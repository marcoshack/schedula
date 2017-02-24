package repository

import (
	"fmt"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/marcoshack/schedula/entity"
)

// JobsInMemoryWithChannels ...
type JobsInMemoryWithChannels struct {
	JobsByID       map[string]*entity.Job
	JobsBySchedule map[int64][]*entity.Job
	requests       chan request
}

type request struct {
	job      *entity.Job
	response chan response
	f        func(*entity.Job, chan response)
}

type response struct {
	job *entity.Job
	err error
}

// NewJobsInMemoryWithChannels ...
func NewJobsInMemoryWithChannels() (*JobsInMemoryWithChannels, error) {
	repo := &JobsInMemoryWithChannels{
		JobsByID:       make(map[string]*entity.Job),
		JobsBySchedule: make(map[int64][]*entity.Job),
		requests:       make(chan request),
	}
	go repo.handle()
	return repo, nil
}

func (r *JobsInMemoryWithChannels) handle() {
	for req := range r.requests {
		req.f(req.job, req.response)
	}
}

func (r *JobsInMemoryWithChannels) add(job *entity.Job, res chan response) {
	if job.ID != "" {
		res <- response{job: job, err: fmt.Errorf("cannot add an existing job")}
		return
	}

	timestamp, err := job.Schedule.NextTimestamp()
	if err != nil {
		res <- response{job: job, err: fmt.Errorf("invalid job schedule: %v", err)}
		return
	}

	job.ID = uuid.New()
	job.Status = entity.JobStatusPending
	r.JobsByID[job.ID] = job

	if r.JobsBySchedule[timestamp] == nil {
		r.JobsBySchedule[timestamp] = make([]*entity.Job, 0)
	}
	r.JobsBySchedule[timestamp] = append(r.JobsBySchedule[timestamp], job)
	res <- response{job: job, err: nil}
}

func (r *JobsInMemoryWithChannels) get(job *entity.Job, res chan response) {
	resJob := r.JobsByID[job.ID]
	if resJob == nil {
		res <- response{job: job, err: fmt.Errorf("job ID=%s not found", job.ID)}
		return
	}
	res <- response{job: resJob, err: nil}
}

func (r *JobsInMemoryWithChannels) cancel(job *entity.Job, res chan response) {
	resJob := r.JobsByID[job.ID]
	if resJob == nil {
		res <- response{job: job, err: fmt.Errorf("job ID=%s not found", job.ID)}
		return
	}
	resJob.Status = entity.JobStatusCanceled
	res <- response{job: resJob, err: nil}
}

func (r *JobsInMemoryWithChannels) execute(req request) (entity.Job, error) {
	req.response = make(chan response, 1)
	r.requests <- req
	res := <-req.response
	if res.err != nil {
		return entity.Job{}, res.err
	}
	return *res.job, nil
}

// Add ...
func (r *JobsInMemoryWithChannels) Add(job entity.Job) (entity.Job, error) {
	return r.execute(request{f: r.add, job: &job})
}

// Get ...
func (r *JobsInMemoryWithChannels) Get(jobID string) (entity.Job, error) {
	return r.execute(request{f: r.get, job: &entity.Job{ID: jobID}})
}

// Remove ...
// TODO implementation pending
func (r *JobsInMemoryWithChannels) Remove(jobID string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Cancel ...
func (r *JobsInMemoryWithChannels) Cancel(jobID string) (entity.Job, error) {
	return r.execute(request{f: r.cancel, job: &entity.Job{ID: jobID}})
}

// List ...
// TODO implementation pending
func (r *JobsInMemoryWithChannels) List(skip int, limit int) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}

// AddExecution ...
// TODO implementation pending
func (r *JobsInMemoryWithChannels) AddExecution(jobID string, date time.Time, status string, message string) (entity.Job, error) {
	return entity.Job{}, nil
}

// Count ...
// TODO implementation pending
func (r *JobsInMemoryWithChannels) Count() int {
	return len(r.JobsByID)
}

// ListBySchedule ...
// TODO implementation pending
func (r *JobsInMemoryWithChannels) ListBySchedule(timestamp int64) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}
