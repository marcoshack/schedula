package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/marcoshack/schedula"
)

const (
	// SchedulaURL is the URL for Schedula server's jobs resource
	SchedulaURL = "http://localhost:8080/jobs/"
)

func main() {
	n := flag.Int("n", 10, "number of callbacks to create")
	flag.Parse()

	server := &http.Server{
		Addr: "127.0.0.1:8088",
	}
	client := &http.Client{}
	callbackTime := time.Now().Add(5 * time.Second)

	jobsCreated := 0
	for i := 1; i <= *n; i++ {
		job := &schedula.Job{
			CallbackURL: fmt.Sprintf("http://127.0.0.1:8088/callback/%d", i),
			Schedule: schedula.JobSchedule{
				Format: "timestamp",
				Value:  fmt.Sprintf("%v", callbackTime.Unix()),
			},
		}

		var body = new(bytes.Buffer)
		encErr := json.NewEncoder(body).Encode(job)
		if encErr != nil {
			log.Printf("ERROR: unable to encode request body: %v", encErr)
			continue
		}

		req, reqErr := http.NewRequest("POST", SchedulaURL, body)
		if reqErr != nil {
			log.Printf("ERROR: failed to create HTTP request: %v", reqErr)
			continue
		}
		req.Header.Set("User-Agent", "schedula")
		req.Header.Set("Content-Type", "application/json")

		res, postErr := client.Do(req)
		if postErr != nil {
			log.Printf("ERROR: failed to send HTTP request: %v", postErr)
			continue
		}

		if res.StatusCode != http.StatusCreated {
			log.Printf("ERROR: invalid response code, expected 201 Created but got %s", res.Status)
			continue
		}

		log.Printf("INFO: callback created: %s", res.Header.Get("Location"))
		jobsCreated++
	}

	http.HandleFunc("/callback/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("INFO: Callback received %s", r.URL.Path)
	})

	if jobsCreated > 0 {
		log.Printf("INFO: %d callbacks created", jobsCreated)
		log.Printf("INFO: Listening for callbacks on %s\n", server.Addr)
		log.Fatal(server.ListenAndServe())
	} else {
		log.Printf("INFO: No jobs were created, terminating.")
	}
}
