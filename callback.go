package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// CallbackExecutor is responsible for executing job's callback
type CallbackExecutor interface {
	Execute(Job) (Job, error)
}

// NewCallbackExecutor returns an instance of CallbackExecutor
func NewCallbackExecutor(repo Repository) (CallbackExecutor, error) {
	return &SynchronousCallbackExecutor{
		repository: repo,
		httpClient: &http.Client{},
	}, nil
}

// SynchronousCallbackExecutor ...
type SynchronousCallbackExecutor struct {
	repository Repository
	httpClient *http.Client
}

// Execute ...
func (s *SynchronousCallbackExecutor) Execute(job Job) (Job, error) {
	var newJobStatus string

	req, err := s.createCallbackRequest(job)
	if err != nil {
		newJobStatus = JobStatusError
		return Job{}, err
	}

	res, err := s.httpClient.Do(req)
	if err != nil {
		newJobStatus = JobStatusError
		return Job{}, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		newJobStatus = JobStatusFail
		return Job{}, fmt.Errorf("invalid callback response: %s", res.Status)
	}

	newJobStatus = JobStatusSuccess
	updatedJob, err := s.repository.SetStatus(job.ID, newJobStatus)
	if err != nil {
		return updatedJob, fmt.Errorf("unable to update job status to '%s': %v", newJobStatus, err)
	}
	
	return updatedJob, nil
}

func (s *SynchronousCallbackExecutor) createCallbackRequest(job Job) (*http.Request, error) {
	var body = new(bytes.Buffer)
	encErr := json.NewEncoder(body).Encode(job)
	if encErr != nil {
		return nil, fmt.Errorf("unable to encode request body: %v", encErr)
	}

	req, err := http.NewRequest("POST", job.CallbackURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "schedula")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
