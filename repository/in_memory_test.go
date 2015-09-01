package repository

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/marcoshack/schedula/entity"
)

func TestNew(t *testing.T) {
	_, err := New("in-memory")
	if err != nil {
		t.Fatalf("unable to initialize repository: %v", err)
	}
}

func TestAdd(t *testing.T) {
	repo, _ := New("in-memory")
	job, err := repo.Add(aJob())
	if err != nil {
		t.Fatalf("unable to add job: %v", err)
	}
	if job.ID == "" {
		t.Fatalf("invalid job ID: %s", job.ID)
	}
}

func TestGet(t *testing.T) {
	repo, _ := New("in-memory")
	newJob, _ := repo.Add(entity.Job{})

	job, _ := repo.Get(newJob.ID)
	if job.ID != newJob.ID {
		log.Print("unable to retrieve job, expected a valid job but got nil")
	}
}

func TestGetNonExistingJob(t *testing.T) {
	repo, _ := New("in-memory")
	job, _ := repo.Get("non-existing-job-id")
	if job.ID != "" {
		t.Fatalf("expected nil but got %s", job)
	}
}

func TestList(t *testing.T) {
	repo, _ := New("in-memory")
	var n = addJobs(repo, 5)
	jobs, _ := repo.List(0, 10)
	if len(jobs) != n {
		t.Fatalf("expected jobs list size to be %d but got %d", n, len(jobs))
	}
}

func TestListWithPagination(t *testing.T) {
	repo, _ := New("in-memory")
	addJobs(repo, 10)
	jobs, _ := repo.List(1, 5)
	if len(jobs) != 5 {
		t.Fatalf("expected jobs list size to be 5, but got %d", len(jobs))
	}
}

func TestRemove(t *testing.T) {
	repo, _ := New("in-memory")
	initialSize := 5
	expectedFinalSize := 4
	var idToRemove string
	for i := 0; i < initialSize; i++ {
		newJob, err := repo.Add(aJob())
		if err != nil {
			t.Fatalf("unable to add job: %v", err)
		}
		if i == (initialSize / 2) {
			idToRemove = newJob.ID
		}
	}
	job, err := repo.Remove(idToRemove)
	if err != nil {
		t.Fatalf("unable to remove job: %v", err)
	}
	if job.ID != idToRemove {
		t.Fatalf("expected removed job to have ID '%s' but got '%s'", idToRemove, job.ID)
	}
	if repo.Count() != expectedFinalSize {
		t.Fatalf("expected repo final size to be %d but got %d", expectedFinalSize, repo.Count())
	}
}

func TestCancel(t *testing.T) {
	repo, _ := New("in-memory")
	job, err := repo.Add(aJob())
	if err != nil {
		t.Fatalf("unable to add job: %v", err)
	}
	repo.Cancel(job.ID)
	assertStatus(t, repo, job.ID, entity.JobStatusCanceled)
}

func TestSetStatus(t *testing.T) {
	repo, _ := New("in-memory")
	job, _ := repo.Add(aJob())
	repo.UpdateStatus(job.ID, entity.JobStatusError)
	assertStatus(t, repo, job.ID, entity.JobStatusError)
}

func ExampleScheduler_List_ordering() {
	repo, _ := New("in-memory")
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

func addJobs(repo Repository, n int) int {
	for i := 0; i < n; i++ {
		repo.Add(entity.Job{BusinessKey: strconv.Itoa(i)})
	}
	return n
}

func assertStatus(t *testing.T, repo Repository, jobID string, expectedStatus string) {
	updatedJob, _ := repo.Get(jobID)
	if updatedJob.Status != expectedStatus {
		t.Fatalf("expected job status was '%s' but got '%s'", expectedStatus, updatedJob.Status)
	}
}

func aJob() entity.Job {
	return entity.Job{
		Schedule: entity.JobSchedule{
			Format: entity.ScheduleFormatTimestamp,
			Value:  "1234567890"},
	}
}
