package repository

import (
	"sync"
	"time"

	"github.com/marcoshack/schedula/entity"

	"code.google.com/p/go-uuid/uuid"
)

// JobsInMemoryWithMutex ...
type JobsInMemoryWithMutex struct {
	sync.RWMutex
	jobsByID       map[string]*entity.Job
	jobsBySchedule map[int64][]*entity.Job
	jobIndexByID   []string
}

// NewJobsInMemoryWithMutex ...
func NewJobsInMemoryWithMutex() (Jobs, error) {
	return &JobsInMemoryWithMutex{
		jobsByID:       make(map[string]*entity.Job),
		jobsBySchedule: make(map[int64][]*entity.Job),
	}, nil
}

// Add ...
func (r *JobsInMemoryWithMutex) Add(job entity.Job) (entity.Job, error) {
	job.ID = uuid.New()
	job.Status = entity.JobStatusPending

	r.Lock()
	defer r.Unlock()

	r.jobIndexByID = append(r.jobIndexByID, job.ID)
	r.jobsByID[job.ID] = &job

	jobTime, err := job.Schedule.NextTimestamp()
	if err != nil {
		return entity.Job{}, err
	}
	if r.jobsBySchedule[jobTime] == nil {
		r.jobsBySchedule[jobTime] = make([]*entity.Job, 0)
	}
	r.jobsBySchedule[jobTime] = append(r.jobsBySchedule[jobTime], &job)

	return job, nil
}

// Get returns the Job associated with the given id or nil if it doensn't exist
func (r *JobsInMemoryWithMutex) Get(id string) (entity.Job, error) {
	job, _ := r.get(id)
	if job != nil {
		return *job, nil
	}
	return entity.Job{}, nil
}

func (r *JobsInMemoryWithMutex) get(id string) (*entity.Job, error) {
	r.RLock()
	defer r.RUnlock()
	return r.jobsByID[id], nil
}

// List returns the list of scheduled jobs
func (r *JobsInMemoryWithMutex) List(skip int, limit int) ([]entity.Job, error) {
	r.RLock()
	defer r.RUnlock()

	// empty result
	if len(r.jobsByID) == 0 || skip > len(r.jobsByID) || limit < 0 {
		return make([]entity.Job, 0), nil
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
	jobs := make([]entity.Job, len(ids))
	for i := 0; i < len(ids); i++ {
		jobs[i] = *r.jobsByID[ids[i]]
	}

	return jobs, nil
}

// Count returns the number of active jobs in the scheduler
func (r *JobsInMemoryWithMutex) Count() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.jobIndexByID)
}

// ListBySchedule returns the list of Jobs scheduled for the given timestamp
func (r *JobsInMemoryWithMutex) ListBySchedule(timestamp int64) ([]entity.Job, error) {
	r.RLock()
	defer r.RUnlock()
	schedList := r.jobsBySchedule[timestamp]

	res := make([]entity.Job, len(schedList))
	for i := 0; i < len(res); i++ {
		res[i] = *schedList[i]
	}

	return res, nil
}

// Remove ...
func (r *JobsInMemoryWithMutex) Remove(jobID string) (entity.Job, error) {
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
	newScheduledJobs := make([]*entity.Job, len(scheduledJobs)-1)
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
func (r *JobsInMemoryWithMutex) Cancel(jobID string) (entity.Job, error) {
	job, err := r.get(jobID)
	if err != nil {
		return entity.Job{}, err
	}
	job.Status = entity.JobStatusCanceled
	return *job, nil
}

// AddExecution ...
func (r *JobsInMemoryWithMutex) AddExecution(jobID string, date time.Time, status string, message string) (entity.Job, error) {
	job, err := r.get(jobID)
	if err != nil {
		return entity.Job{}, err
	}
	job.Executions = append(job.Executions, entity.JobExecution{Timestamp: date.Unix(), Status: status, Message: message})
	job.Status = status
	return *job, nil
}
