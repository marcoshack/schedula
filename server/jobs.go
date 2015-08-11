package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/marcoshack/schedula"
)

const (
	// MaxPageSize is the maximum number of items for listing resources
	MaxPageSize = 100
)

// JobsHandler is a HTTP handler to retrieve and manipulate jobs
type JobsHandler struct {
	Path      string
	scheduler schedula.Scheduler
}

func (h *JobsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		h.List(w, r)
	case "POST":
		h.Create(w, r)
	}
}

// List all jobs in JSON format
func (h *JobsHandler) List(w http.ResponseWriter, r *http.Request) {
	p := make(map[string]interface{})
	p["skip"] = ParseIntParam(r, "skip", 0)
	p["limit"] = ParseIntParam(r, "limit", MaxPageSize)
	log.Printf("jobs: listing jobs (params: %v)", p)

	jobs, err := h.scheduler.List(p["skip"].(int), p["limit"].(int))
	if err != nil {
		ErrorResponse(err, w, http.StatusInternalServerError)
		return
	}

	var resBuf = new(bytes.Buffer)
	encErr := json.NewEncoder(resBuf).Encode(jobs)
	if encErr != nil {
		ErrorResponse(encErr, w, http.StatusInternalServerError)
	}
	w.Header().Add("Page-Count", strconv.Itoa(len(jobs)))
	w.Header().Add("Total-Count", strconv.Itoa(h.scheduler.Size()))
	w.Write(resBuf.Bytes())
}

// Create a job from a JSON representation
func (h *JobsHandler) Create(w http.ResponseWriter, r *http.Request) {
	job, err := ParseJob(r)
	if err != nil {
		log.Printf("jobs: error parsing job: %s", err)
		ErrorResponse(err, w, http.StatusBadRequest)
		return
	}

	id, err := h.scheduler.Add(job)
	if err != nil {
		log.Printf("jobs: error scheduling job: %s", err)
		ErrorResponse(err, w, http.StatusInternalServerError)
		return
	}

	log.Printf("jobs: job created: %s", job)
	w.Header().Add("Location", fmt.Sprintf("%s%s", h.Path, id))
	w.WriteHeader(http.StatusCreated)
}
