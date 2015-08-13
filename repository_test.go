package schedula

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestNewRepository(t *testing.T) {
	_, err := NewRepository()
	if err != nil {
		t.Fatalf("unable to initialize repository: %v", err)
	}
}

func TestAddAndRetrieveJobs(t *testing.T) {
	repo, _ := NewRepository()
	newJob, e := repo.Add(Job{
		CallbackURL: "http://example.com",
		Schedule: JobSchedule{
			Format: ScheduleFormatTimestamp,
			Value:  strconv.FormatInt(time.Now().Unix(), 10),
		},
	})

	if e != nil {
		t.Fatalf("failed adding a new job to scheduler: %s", e)
	}
	if repo.Count() != 1 {
		t.Fatalf("invalid scheduler size, expected 1 but got %d", repo.Count())
	}
	job, _ := repo.Get(newJob.ID)
	if job.ID != newJob.ID {
		log.Print("failed retrieving job, expected a valid job but got nil")
	}
}

func TestRetrieveANonExistingJob(t *testing.T) {
	repo, _ := NewRepository()
	job, _ := repo.Get("non-existing-job-id")
	if job.ID != "" {
		t.Fatalf("expected nil but got %s", job)
	}
}

func TestList(t *testing.T) {
	repo, _ := NewRepository()
	var n = addJobs(repo, 5)
	jobs, _ := repo.List(0, 10)
	if len(jobs) != n {
		t.Fatalf("expected jobs list size to be %d but got %d", n, len(jobs))
	}
}

func TestListWithPagination(t *testing.T) {
	repo, _ := NewRepository()
	addJobs(repo, 10)
	jobs, _ := repo.List(1, 5)
	if len(jobs) != 5 {
		t.Fatalf("expected jobs list size to be 5, but got %d", len(jobs))
	}
}

func ExampleScheduler_List_ordering() {
	repo, _ := NewRepository()
	n := 10
	addJobs(repo, n)
	jobs, _ := repo.List(0, n)
	keys := make([]string, n)
	for j := range jobs {
		keys[j] = jobs[j].BusinessKey
	}
	fmt.Println(keys)
	// Output:
	// [0 1 2 3 4 5 6 7 8 9]
}

func aJobWithBusinessKey(businessKey string) Job {
	return Job{
		CallbackURL: "http://example.com",
		BusinessKey: businessKey,
		Schedule: JobSchedule{
			Format: ScheduleFormatTimestamp,
			Value:  strconv.FormatInt(time.Now().Unix(), 10),
		},
	}
}

func addJobs(repo Repository, n int) int {
	for i := 0; i < n; i++ {
		repo.Add(aJobWithBusinessKey(strconv.Itoa(i)))
	}
	return n
}
