package callback

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/marcoshack/schedula/entity"
	"github.com/marcoshack/schedula/repository"
)

// Executor is responsible for executing job's callback
type Executor interface {
	Execute(entity.Job) error
}

// NewExecutor returns an instance of Executor
func NewExecutor() (Executor, error) {
	return &SynchronousExecutor{
		httpClient: &http.Client{},
	}, nil
}

// SynchronousExecutor ...
type SynchronousExecutor struct {
	repository repository.Jobs
	httpClient *http.Client
}

// Execute ...
func (s *SynchronousExecutor) Execute(job entity.Job) error {
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

func (s *SynchronousExecutor) createCallbackRequest(job entity.Job) (*http.Request, error) {
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
