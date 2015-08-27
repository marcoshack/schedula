package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// CallbackExecutor is responsible for executing job's callback
type CallbackExecutor interface {
	Execute(Job) error
}

// NewCallbackExecutor returns an instance of CallbackExecutor
func NewCallbackExecutor() (CallbackExecutor, error) {
	return &SynchronousCallbackExecutor{
		httpClient: &http.Client{},
	}, nil
}

// SynchronousCallbackExecutor ...
type SynchronousCallbackExecutor struct {
	repository Repository
	httpClient *http.Client
}

// Execute ...
func (s *SynchronousCallbackExecutor) Execute(job Job) error {
	req, err := s.createCallbackRequest(job)
	if err != nil {
		return err
	}

	res, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("invalid callback response, expect 200 OK or 202 Accepted but got %s", res.Status)
	}

	return nil
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
