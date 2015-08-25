package schedula

import "testing"

func createScheduler(t *testing.T) (Scheduler, *RepositoryMock) {
	r := NewRepositoryMock()
	s, e := InitAndStartScheduler(r, SchedulerConfig{})
	if e != nil {
		t.Fatalf("failed to initialize scheduler: %s", e)
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
type RepositoryMock struct {
	counters map[string]int
}

func NewRepositoryMock() *RepositoryMock {
	return &RepositoryMock{counters: make(map[string]int)}
}

func (r *RepositoryMock) Counter(method string) int {
	return r.counters[method]
}

func (r *RepositoryMock) Add(job Job) (Job, error) {
	r.inc("Add")
	return job, nil
}

func (r *RepositoryMock) Get(id string) (Job, error) {
	r.inc("Get")
	return Job{ID: id}, nil
}

func (r *RepositoryMock) List(skip int, limit int) ([]Job, error) {
	r.inc("List")
	return make([]Job, 0), nil
}

func (r *RepositoryMock) Count() int {
	r.inc("Count")
	return r.counters["Count"]
}

func (r *RepositoryMock) ListBySchedule(timestamp int64) ([]*Job, error) {
	return make([]*Job, 0), nil
}

func (r *RepositoryMock) Remove(jobID string) (Job, error) {
	r.inc("Remove")
	return Job{ID: jobID}, nil
}

func (r *RepositoryMock) Cancel(jobID string) (Job, error) {
	r.inc("Cancel")
	return Job{ID: jobID}, nil
}

func (r *RepositoryMock) inc(method string) {
	r.counters[method]++
}
