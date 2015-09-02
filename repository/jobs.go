package repository

import (
	"fmt"
	"time"

	"github.com/marcoshack/schedula/entity"
)

// Jobs ...
type Jobs interface {
	Add(entity.Job) (entity.Job, error)
	Get(id string) (entity.Job, error)
	List(skip int, limit int) ([]entity.Job, error)
	Remove(jobID string) (entity.Job, error)
	Cancel(jobID string) (entity.Job, error)
	AddExecution(jobID string, date time.Time, status string, message string) (entity.Job, error)
	Count() int
	ListBySchedule(timestamp int64) ([]entity.Job, error)
}

// New creates a repository instance of the given type.
func New(repoType string) (Jobs, error) {
	switch repoType {
	case "in-memory", "in-memory-mutex": // in-memory default, for now
		return NewJobsInMemoryWithMutex()
	case "in-memory-ch":
		return NewJobsInMemoryWithChannels()
	case "redis":
		return NewJobsRedis()
	case "mysql":
		return NewJobsMySQL()
	}
	return nil, fmt.Errorf("invalid repository type: '%s'", repoType)
}
