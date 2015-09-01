package repository

import (
	"fmt"

	"github.com/marcoshack/schedula/entity"
)

// Repository ...
type Repository interface {
	Add(entity.Job) (entity.Job, error)
	Get(id string) (entity.Job, error)
	List(skip int, limit int) ([]entity.Job, error)
	Remove(jobID string) (entity.Job, error)
	Cancel(jobID string) (entity.Job, error)
	UpdateStatus(jobID string, status string) (entity.Job, error)
	Count() int
	ListBySchedule(timestamp int64) ([]entity.Job, error)
}

// New creates a repository instance of the given type.
func New(repoType string) (Repository, error) {
	switch repoType {
	case "in-memory":
		return &InMemoryJobRepository{
			jobsByID:       make(map[string]*entity.Job),
			jobsBySchedule: make(map[int64][]*entity.Job),
		}, nil
	}
	return nil, fmt.Errorf("invalid repository type: '%s'", repoType)
}