package main

import (
	"sync"

	"github.com/marcoshack/schedula/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
)

// Repository ...
type Repository interface {
	Add(Job) (Job, error)
	Get(id string) (Job, error)
	List(skip int, limit int) ([]Job, error)
	Remove(jobID string) (Job, error)
	Cancel(jobID string) (Job, error)
	UpdateStatus(jobID string, status string) (Job, error)
	Count() int
	ListBySchedule(timestamp int64) ([]Job, error)
}

// NewRepository ...
func NewRepository() (Repository, error) {
	return &InMemoryJobRepository{
		jobsByID:       make(map[string]*Job),
		jobsBySchedule: make(map[int64][]*Job),
	}, nil
}

// InMemoryJobRepository ...
type InMemoryJobRepository struct {
	sync.RWMutex
	jobsByID       map[string]*Job
	jobsBySchedule map[int64][]*Job
	jobIndexByID   []string
}

// Add ...
func (r *InMemoryJobRepository) Add(job Job) (Job, error) {
	job.ID = uuid.New()
	job.Status = JobStatusPending

	r.Lock()
	defer r.Unlock()

	r.jobIndexByID = append(r.jobIndexByID, job.ID)
	r.jobsByID[job.ID] = &job

	jobTime, err := job.Schedule.NextTimestamp()
	if err != nil {
		return Job{}, err
	}
	if r.jobsBySchedule[jobTime] == nil {
		r.jobsBySchedule[jobTime] = make([]*Job, 0)
	}
	r.jobsBySchedule[jobTime] = append(r.jobsBySchedule[jobTime], &job)

	return job, nil
}

// Get returns the Job associated with the given id or nil if it doensn't exist
func (r *InMemoryJobRepository) Get(id string) (Job, error) {
	job, _ := r.get(id)
	if job != nil {
		return *job, nil
	}
	return Job{}, nil
}

func (r *InMemoryJobRepository) get(id string) (*Job, error) {
	r.RLock()
	defer r.RUnlock()
	return r.jobsByID[id], nil
}

// List returns the list of scheduled jobs
func (r *InMemoryJobRepository) List(skip int, limit int) ([]Job, error) {
	r.RLock()
	defer r.RUnlock()

	// empty result
	if len(r.jobsByID) == 0 || skip > len(r.jobsByID) || limit < 0 {
		return make([]Job, 0), nil
	}

	start := skip
	if start > len(r.jobIndexByID) {
		skip = len(r.jobIndexByID)
	}
	end := skip + limit
	if end > len(r.jobIndexByID) {
		end = len(r.jobIndexByID)
	}

	ids := r.jobIndexByID[start:end]
	jobs := make([]Job, len(ids))
	for i := 0; i < len(ids); i++ {
		jobs[i] = *r.jobsByID[ids[i]]
	}

	return jobs, nil
}

// Count returns the number of active jobs in the scheduler
func (r *InMemoryJobRepository) Count() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.jobIndexByID)
}

// ListBySchedule returns the list of Jobs scheduled for the given timestamp
func (r *InMemoryJobRepository) ListBySchedule(timestamp int64) ([]Job, error) {
	r.RLock()
	defer r.RUnlock()
	schedList := r.jobsBySchedule[timestamp]

	res := make([]Job, len(schedList))
	for i := 0; i < len(res); i++ {
		res[i] = *schedList[i]
	}

	return res, nil
}

// Remove ...
func (r *InMemoryJobRepository) Remove(jobID string) (Job, error) {
	job, err := r.get(jobID)
	if err != nil {
		return *job, err
	}

	jobTimestamp, err := job.Schedule.NextTimestamp()
	if err != nil {
		return *job, err
	}

	r.Lock()
	defer r.Unlock()

	// remove from r.jobsByID
	delete(r.jobsByID, jobID)

	// rebuild r.jobIndexByID
	// TODO use append to rebuild
	newJobIndex := make([]string, len(r.jobIndexByID)-1)
	for i, j := 0, 0; i < len(newJobIndex); i, j = i+1, j+1 {
		if r.jobIndexByID[i] == jobID {
			j++
		}
		newJobIndex[i] = r.jobIndexByID[j]
	}
	r.jobIndexByID = newJobIndex

	// remove from r.jobsBySchedule
	// TODO use append to rebuild
	scheduledJobs := r.jobsBySchedule[jobTimestamp]
	newScheduledJobs := make([]*Job, len(scheduledJobs)-1)
	for i, j := 0, 0; i < len(newScheduledJobs); i, j = i+1, j+1 {
		if scheduledJobs[i].ID == jobID {
			scheduledJobs[j] = nil
			j++
		}
		newScheduledJobs[i] = scheduledJobs[j]
	}
	r.jobsBySchedule[jobTimestamp] = newScheduledJobs

	return *job, nil
}

// Cancel changes job status to JobStatusCanceled
func (r *InMemoryJobRepository) Cancel(jobID string) (Job, error) {
	job, err := r.get(jobID)
	if err != nil {
		return Job{}, err
	}
	job.Status = JobStatusCanceled
	return *job, nil
}

// UpdateStatus changes the status of the job identified by the given ID
func (r *InMemoryJobRepository) UpdateStatus(jobID string, status string) (Job, error) {
	job, err := r.get(jobID)
	if err != nil {
		return Job{}, err
	}
	job.Status = status
	return *job, nil
}
