package schedula

import (
	"log"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

// Repository ...
type Repository interface {
	Add(Job) (Job, error)
	Get(id string) (Job, error)
	List(skip int, limit int) ([]Job, error)
	Count() int
	ListBySchedule(timestamp int64) ([]*Job, error)
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

	jobTimestamp, err := job.Schedule.NextTimestamp()
	if err != nil {
		return Job{}, err
	}
	if r.jobsBySchedule[jobTimestamp] == nil {
		r.jobsBySchedule[jobTimestamp] = make([]*Job, 0)
	}
	r.jobsBySchedule[jobTimestamp] = append(r.jobsBySchedule[jobTimestamp], &job)

	return job, nil
}

// Get returns the Job associated with the given id or nil if it doensn't exist
func (r *InMemoryJobRepository) Get(id string) (Job, error) {
	r.RLock()
	defer r.RUnlock()
	if r.jobsByID[id] != nil {
		return *r.jobsByID[id], nil
	}
	return Job{}, nil
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
func (r *InMemoryJobRepository) ListBySchedule(timestamp int64) ([]*Job, error) {
	lockRequest := time.Now()
	r.RLock()
	lockAcquired := time.Now()
	schedList := r.jobsBySchedule[timestamp]
	r.RUnlock()
	lockReleased := time.Now()
	lockDuration := lockReleased.Sub(lockAcquired)
	lockAcquiring := lockAcquired.Sub(lockRequest)

	if lockDuration > 1*time.Millisecond || lockAcquiring > 1*time.Millisecond {
		log.Printf("scheduler: jobs lock duration: %fs, acquiring: %fs", lockDuration.Seconds(), lockAcquiring.Seconds())
	}
	return schedList, nil
}
