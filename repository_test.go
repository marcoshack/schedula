package main

import (
	"fmt"
	"log"
	"strconv"
	"testing"
)

func TestNewRepository(t *testing.T) {
	_, err := NewRepository()
	if err != nil {
		t.Fatalf("unable to initialize repository: %v", err)
	}
}

func TestAdd(t *testing.T) {
	repo, _ := NewRepository()
	job, err := repo.Add(aJob())
	if err != nil {
		t.Fatalf("unable to add job: %v", err)
	}
	if job.ID == "" {
		t.Fatalf("invalid job ID: %s", job.ID)
	}
}

func TestGet(t *testing.T) {
	repo, _ := NewRepository()
	newJob, _ := repo.Add(aJob())

	job, _ := repo.Get(newJob.ID)
	if job.ID != newJob.ID {
		log.Print("unable to retrieve job, expected a valid job but got nil")
	}
}

func TestGetNonExistingJob(t *testing.T) {
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

func TestRemove(t *testing.T) {
	repo, _ := NewRepository()
	initialSize := 5
	expectedFinalSize := 4
	var toRemove string
	for i := 0; i < initialSize; i++ {
		newJob, _ := repo.Add(aJob())
		if i == 2 {
			toRemove = newJob.ID
		}
	}
	job, err := repo.Remove(toRemove)
	if err != nil {
		t.Fatalf("unable to remove job: %v", err)
	}
	if job.ID != toRemove {
		t.Fatalf("expected removed job to have ID '%s' but got '%s'", toRemove, job.ID)
	}
	if repo.Count() != expectedFinalSize {
		t.Fatalf("expected repo final size to be %d but got %d", expectedFinalSize, repo.Count())
	}
}

func TestCancel(t *testing.T) {
	repo, _ := NewRepository()
	job, _ := repo.Add(aJob())
	repo.Cancel(job.ID)
	updatedJob, _ := repo.Get(job.ID)
	if updatedJob.Status != JobStatusCanceled {
		t.Fatalf("expected job status was '%s' but got '%s'", JobStatusCanceled, updatedJob.Status)
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

func addJobs(repo Repository, n int) int {
	for i := 0; i < n; i++ {
		repo.Add(aJobWithBusinessKey(strconv.Itoa(i)))
	}
	return n
}
