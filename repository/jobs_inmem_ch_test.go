package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/marcoshack/schedula/entity"
)

func Test_JobsInMemoryWithChannels_Add(t *testing.T) {
	repo, _ := NewJobsInMemoryWithChannels()
	job, err := repo.Add(entity.Job{
		Schedule: entity.JobSchedule{
			Format: "timestamp",
			Value:  fmt.Sprintf("%v", time.Now().Unix()),
		},
	})

	if err != nil {
		t.Fatalf("error creating job: %v", err)
	}

	if job.ID == "" {
		t.Fatalf("expected job ID not to be blank")
	}

	if repo.JobsByID[job.ID] == nil {
		t.Fatalf("expected job to be added to JobsByID")
	}

	if timestamp, _ := job.Schedule.NextTimestamp(); repo.JobsBySchedule[timestamp] == nil {
		t.Fatalf("expected job to be added to JobsBySchedule")
	}
}

func Test_JobsInMemoryWithChannels_Get(t *testing.T) {
	repo, _ := NewJobsInMemoryWithChannels()
	repo.JobsByID["foo"] = &entity.Job{ID: "foo"}
	job, err := repo.Get("foo")
	if err != nil {
		t.Fatalf("error retrieving job from repository: %v", err)
	}
	if job.ID != "foo" {
		t.Fatalf("expected job ID to be 'foo' but bot '%s'", job.ID)
	}
}

func Test_JobsInMemoryWithChannels_GetNotFound(t *testing.T) {
	repo, _ := NewJobsInMemoryWithChannels()
	job, err := repo.Get("foo")
	if err == nil {
		t.Fatalf("expected to receive an error for a non existing job ID")
	}
	if job.ID != "" {
		t.Fatalf("expected job ID to be blank but got '%s'", job.ID)
	}
}

func Test_JobsInMemoryWithChannels_Cancel(t *testing.T) {
	repo, _ := NewJobsInMemoryWithChannels()
	repo.JobsByID["foo"] = &entity.Job{ID: "foo", Status: entity.JobStatusPending}
	job, err := repo.Cancel("foo")
	if err != nil {
		t.Fatalf("error canceling job: %v", err)
	}
	if job.Status != entity.JobStatusCanceled {
		t.Fatalf("expected job status to be '%s' but got '%s'", entity.JobStatusCanceled, job.Status)
	}
	if repo.JobsByID["foo"].Status != entity.JobStatusCanceled {
		t.Fatalf("expected job status in JobsByID to be '%s' but got '%s'", entity.JobStatusCanceled, job.Status)
	}
}
