package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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
	skip := ParseIntParam(r, "skip", 0)
	limit := ParseIntParam(r, "limit", MaxPageSize)

	jobs, err := h.scheduler.List(skip, limit)
	if err != nil {
		ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	var resBuf = new(bytes.Buffer)
	encErr := json.NewEncoder(resBuf).Encode(jobs)
	if encErr != nil {
		ErrorResponse(w, encErr, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Page-Count", strconv.Itoa(len(jobs)))
	w.Header().Add("Total-Count", strconv.Itoa(h.scheduler.Count()))
	w.Write(resBuf.Bytes())
}

// Create a job from a JSON representation
func (h *JobsHandler) Create(w http.ResponseWriter, r *http.Request) {
	job, err := ParseJob(r)
	if err != nil {
		log.Printf("jobs: error parsing job: %s", err)
		ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	newJob, err := h.scheduler.Add(job)
	if err != nil {
		log.Printf("jobs: error scheduling job: %s", err)
		ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", fmt.Sprintf("%s%s", h.Path, newJob.ID))
	w.WriteHeader(http.StatusCreated)
}

// Find retrieves the job specified by the 'id' path parameter
func (h *JobsHandler) Find(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	job, err := h.scheduler.Get(id)

	if err != nil {
		ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	if job.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var resBuf = new(bytes.Buffer)
	encErr := json.NewEncoder(resBuf).Encode(job)
	if encErr != nil {
		ErrorResponse(w, encErr, http.StatusInternalServerError)
		return
	}
	w.Write(resBuf.Bytes())
}
