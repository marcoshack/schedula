package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/marcoshack/schedula"
)

// ErrorResponse ...
func ErrorResponse(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
}

// ParseJob ...
func ParseJob(r *http.Request) (schedula.Job, error) {
	var job schedula.Job
	dec := json.NewDecoder(r.Body)
	return job, dec.Decode(&job)
}

// ParseIntParam ...
func ParseIntParam(r *http.Request, name string, defaultValue int) int {
	res := defaultValue
	if r.URL.Query().Get(name) != "" {
		value, err := strconv.Atoi(r.URL.Query().Get(name))
		if err == nil {
			res = value
		}
	}
	return res
}
