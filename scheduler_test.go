package schedula

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestSchedulerInitialization(t *testing.T) {
	s, _ := createScheduler(t)
	if s.Size() != 0 {
		t.Fatalf("expected scheduler size to be 0 but got %d", s.Size())
	}
}

func TestAddAndRetrieveJobs(t *testing.T) {
	s, _ := createScheduler(t)
	newJob, e := s.Add(Job{
		CallbackURL: "http://example.com",
		Schedule: JobSchedule{
			Format: ScheduleFormatTimestamp,
			Value:  strconv.FormatInt(time.Now().Unix(), 10),
		},
	})

	if e != nil {
		t.Fatalf("failed adding a new job to scheduler: %s", e)
	}
	if s.Size() != 1 {
		t.Fatalf("invalid scheduler size, expected 1 but got %d", s.Size())
	}
	job, _ := s.Get(newJob.ID)
	if job.ID != newJob.ID {
		log.Print("scheduler_test: failed retrieving job, expected a valid job but got nil")
	}
}

func TestRetrieveANonExistingJob(t *testing.T) {
	s, _ := createScheduler(t)
	job, _ := s.Get("non-existing-job-id")
	if job.ID != "" {
		t.Fatalf("expected nil but got %s", job)
	}
}

func TestList(t *testing.T) {
	s, _ := createScheduler(t)
	var n = addJobs(s, 5)
	jobs, _ := s.List(0, 10)
	if len(jobs) != n {
		t.Fatalf("expected jobs list size to be %d but got %d", n, len(jobs))
	}
}

func TestListWithPagination(t *testing.T) {
	s, _ := createScheduler(t)
	addJobs(s, 10)
	jobs, _ := s.List(1, 5)
	if len(jobs) != 5 {
		t.Fatalf("expected jobs list size to be 5, but got %d", len(jobs))
	}
}

func ExampleScheduler_List_ordering() {
	s, _ := createScheduler(&testing.T{})
	n := 10
	addJobs(s, n)
	jobs, _ := s.List(0, n)
	keys := make([]string, n)
	for j := range jobs {
		keys[j] = jobs[j].BusinessKey
	}
	fmt.Println(keys)
	// Output:
	// [0 1 2 3 4 5 6 7 8 9]
}

func createScheduler(t *testing.T) (Scheduler, error) {
	s, e := StartScheduler(TypeInMemory, map[string]interface{}{"no-tick-log": true})
	if e != nil {
		t.Fatalf("failed to initialize scheduler: %s", e)
	}
	return s, nil
}

func addJobs(s Scheduler, n int) int {
	for i := 0; i < n; i++ {
		s.Add(Job{BusinessKey: strconv.Itoa(i)})
	}
	return n
}
