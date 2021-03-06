package entity

import (
	"fmt"
	"strconv"
)

const (
	// ScheduleFormatTimestamp ...
	ScheduleFormatTimestamp = "timestamp"

	// JobStatusSuccess ...
	JobStatusSuccess = "success"

	// JobStatusPending ...
	JobStatusPending = "pending"

	// JobStatusError ...
	JobStatusError = "error"

	// JobStatusCanceled ...
	JobStatusCanceled = "canceled"
)

// Job ...
type Job struct {
	ID          string            `json:"id"`
	ClientKey   string            `json:"clientKey"`
	CallbackURL string            `json:"callbackURL"`
	Data        map[string]string `json:"data"`
	Schedule    JobSchedule       `json:"schedule"`
	Status      string            `json:"status"`
	Executions  []JobExecution    `json:"executions"`
}

// IsExecutable returns whether the job callback should be executed
func (j *Job) IsExecutable() bool {
	return j.Status != JobStatusSuccess && j.Status != JobStatusCanceled
}

// JobSchedule ...
type JobSchedule struct {
	Format string `json:"format"`
	Value  string `json:"value"`
}

// JobExecution ...
type JobExecution struct {
	Timestamp int64
	Status    string
	Message   string
}

// IsValid checks the JobSchedule values and retrun
func (s *JobSchedule) IsValid() bool {
	_, err := s.NextTimestamp()
	if err != nil {
		return false
	}
	return true
}

// NextTimestamp returns the next timestamp (epoch) this schedule occurs.
func (s *JobSchedule) NextTimestamp() (int64, error) {
	switch s.Format {
	case ScheduleFormatTimestamp:
		timestamp, err := strconv.ParseInt(s.Value, 0, 64)
		if err != nil {
			return 0, err
		}
		return timestamp, nil
	}
	return 0, fmt.Errorf("invalid job schedule format: '%s'", s.Format)
}
