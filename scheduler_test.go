package schedula

import "testing"

func TestSchedulerAdd(t *testing.T) {
	s, r := createScheduler(t)
	s.Add(Job{})
	assertReposityCall("Add", 1, r, t)
}

func TestSchedulerList(t *testing.T) {
	s, r := createScheduler(t)
	s.List(0, 10)
	assertReposityCall("List", 1, r, t)
}

func TestSchedulerGet(t *testing.T) {
	s, r := createScheduler(t)
	s.Get("foo")
	assertReposityCall("Get", 1, r, t)
}

func TestSchedulerCount(t *testing.T) {
	s, r := createScheduler(t)
	s.Count()
	assertReposityCall("Count", 1, r, t)
}

func createScheduler(t *testing.T) (Scheduler, *RepositoryMock) {
	r := NewRepositoryMock()
	s, e := InitAndStartScheduler(r, map[string]interface{}{"no-tick-log": true})
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

func (r *RepositoryMock) inc(method string) {
	r.counters[method]++
}
