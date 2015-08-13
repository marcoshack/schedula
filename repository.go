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
		id:    make(map[string]*Job),
		sched: make(map[int64][]*Job),
	}, nil
}

// InMemoryJobRepository ...
type InMemoryJobRepository struct {
	sync.RWMutex
	id    map[string]*Job
	sched map[int64][]*Job
	list  []Job
}

// Add ...
func (r *InMemoryJobRepository) Add(job Job) (Job, error) {
	job.ID = uuid.New()
	job.Status = JobStatusPending

	r.Lock()
	defer r.Unlock()

	// add to job list
	r.list = append(r.list, job)
	jobAddr := &r.list[len(r.list)-1]

	// add to jobs map by ID
	r.id[job.ID] = jobAddr

	// add to jobs map by schedule timetamp
	jobTimestamp, err := job.Schedule.NextTimestamp()
	if err != nil {
		return Job{}, err
	}
	if r.sched[jobTimestamp] == nil {
		r.sched[jobTimestamp] = make([]*Job, 0)
	}
	r.sched[jobTimestamp] = append(r.sched[jobTimestamp], jobAddr)

	return job, nil
}

// Get returns the Job associated with the given id or nil if it doensn't exist
func (r *InMemoryJobRepository) Get(id string) (Job, error) {
	r.RLock()
	defer r.RUnlock()
	if r.id[id] != nil {
		return *r.id[id], nil
	}
	return Job{}, nil
}

// List returns the list of scheduled jobs
func (r *InMemoryJobRepository) List(skip int, limit int) ([]Job, error) {
	r.RLock()
	defer r.RUnlock()

	if len(r.id) == 0 || skip > len(r.id) || limit < 0 {
		return make([]Job, 0), nil
	}

	start := skip
	if start > len(r.list) {
		skip = len(r.list)
	}
	end := skip + limit
	if end > len(r.list) {
		end = len(r.list)
	}

	return r.list[start:end], nil
}

// Count returns the number of active jobs in the scheduler
func (r *InMemoryJobRepository) Count() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.list)
}

// ListBySchedule returns the list of Jobs scheduled for the given timestamp
func (r *InMemoryJobRepository) ListBySchedule(timestamp int64) ([]*Job, error) {
	lockRequest := time.Now()
	r.RLock()
	lockAcquired := time.Now()
	schedList := r.sched[timestamp]
	r.RUnlock()
	lockReleased := time.Now()
	lockDuration := lockReleased.Sub(lockAcquired)
	lockAcquiring := lockAcquired.Sub(lockRequest)

	if lockDuration > 1*time.Millisecond || lockAcquiring > 1*time.Millisecond {
		log.Printf("scheduler: jobs lock duration: %fs, acquiring: %fs", lockDuration.Seconds(), lockAcquiring.Seconds())
	}
	return schedList, nil
}
