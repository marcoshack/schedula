package scheduler

import (
	"testing"

	"github.com/marcoshack/schedula/entity"
)

func createScheduler(t *testing.T) (Scheduler, *RepositoryMock) {
	r := &RepositoryMock{}
	e := &CallbackExecutorMock{}
	s, err := StartNew("in-memory", r, e, Config{})
	if err != nil {
		t.Fatalf("failed to initialize scheduler: %v", err)
	}
	return s, r
}

func assertReposityCall(method string, count int, r *RepositoryMock, t *testing.T) {
	if r.Counter(method) != 1 {
		t.Fatalf("expected 1 call to repository but got %d", r.Counter(method))
	}
}

//
// TODO replace with a mock framework like golang/mock
//
type CounterMock struct {
	counters map[string]int
}

func (m *CounterMock) Counter(method string) int {
	m.assertInit()
	return m.counters[method]
}

func (m *CounterMock) Inc(method string) {
	m.assertInit()
	m.counters[method]++
}

func (m *CounterMock) assertInit() {
	if m.counters == nil {
		m.counters = make(map[string]int)
	}
}

type RepositoryMock struct {
	CounterMock
}

func (r *RepositoryMock) Counter(method string) int {
	return r.counters[method]
}

func (r *RepositoryMock) Add(job entity.Job) (entity.Job, error) {
	r.Inc("Add")
	return job, nil
}

func (r *RepositoryMock) Get(id string) (entity.Job, error) {
	r.Inc("Get")
	return entity.Job{ID: id}, nil
}

func (r *RepositoryMock) List(skip int, limit int) ([]entity.Job, error) {
	r.Inc("List")
	return make([]entity.Job, 0), nil
}

func (r *RepositoryMock) Count() int {
	r.Inc("Count")
	return r.counters["Count"]
}

func (r *RepositoryMock) ListBySchedule(timestamp int64) ([]entity.Job, error) {
	return make([]entity.Job, 0), nil
}

func (r *RepositoryMock) Remove(jobID string) (entity.Job, error) {
	r.Inc("Remove")
	return entity.Job{ID: jobID}, nil
}

func (r *RepositoryMock) Cancel(jobID string) (entity.Job, error) {
	r.Inc("Cancel")
	return entity.Job{ID: jobID}, nil
}

func (r *RepositoryMock) UpdateStatus(jobID string, status string) (entity.Job, error) {
	r.Inc("SetStatus")
	return entity.Job{ID: jobID}, nil
}

type CallbackExecutorMock struct {
	CounterMock
}

func (e *CallbackExecutorMock) Execute(job entity.Job) error {
	e.Inc("Execute")
	return nil
}
