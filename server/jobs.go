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
	skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	log.Printf("jobs: listing jobs (skip=%d, limit=%d)", skip, limit)

	jobs, err := h.scheduler.List(skip, limit)
	if err != nil {
		h.errorResponse(err, w, http.StatusInternalServerError)
		return
	}

	var resBuf = new(bytes.Buffer)
	encErr := json.NewEncoder(resBuf).Encode(jobs)
	if encErr != nil {
		h.errorResponse(encErr, w, http.StatusInternalServerError)
	}
	w.Header().Add("Total-Count", strconv.Itoa(len(jobs)))
	w.Write(resBuf.Bytes())
}

// Create a job from a JSON representation
func (h *JobsHandler) Create(w http.ResponseWriter, r *http.Request) {
	job, err := h.parseJob(r)
	if err != nil {
		log.Printf("jobs: error parsing job: %s", err)
		h.errorResponse(err, w, http.StatusBadRequest)
		return
	}

	id, err := h.scheduler.Add(job)
	if err != nil {
		log.Printf("jobs: error scheduling job: %s", err)
		h.errorResponse(err, w, http.StatusInternalServerError)
		return
	}

	log.Printf("jobs: job created: %s", job)
	w.Header().Add("Location", fmt.Sprintf("%s%s", h.Path, id))
	w.WriteHeader(http.StatusCreated)
}

func (h *JobsHandler) errorResponse(err error, w http.ResponseWriter, status int) {
	fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
	w.WriteHeader(status)
}

func (h *JobsHandler) parseJob(r *http.Request) (schedula.Job, error) {
	var job schedula.Job
	dec := json.NewDecoder(r.Body)
	return job, dec.Decode(&job)
}
