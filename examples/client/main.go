package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Job ...
type Job struct {
	CallbackURL string      `json:"callbackURL"`
	Schedule    JobSchedule `json:"schedule"`
}

// JobSchedule ...
type JobSchedule struct {
	Format string `json:"format"`
	Value  string `json:"value"`
}

func main() {
	numberOfCalbacks := flag.Int("n", 10, "`number` of callbacks to create")
	serverPort := flag.Int("p", 8088, "HTTP `port` number to listen for callbacks")
	callbackTimeDelay := flag.Int("d", 5, "delay in `seconds` to create callbacks")
	serverBaseURL := flag.String("s", "http://localhost:8080/", "Schedula server base `URL`")
	flag.Parse()

	callbackDelayDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", *callbackTimeDelay))
	callbackTime := time.Now().Add(callbackDelayDuration)
	jobsURL := fmt.Sprintf("%sjobs/", *serverBaseURL)

	client := &http.Client{}

	_, err := client.Head(jobsURL)
	if err != nil {
		log.Fatalf("ERROR: schedula server is unavailable: %v", err)
	}

	jobsCreated := 0
	for i := 1; i <= *numberOfCalbacks; i++ {
		job := &Job{
			CallbackURL: fmt.Sprintf("http://127.0.0.1:%d/callback/%d", *serverPort, i),
			Schedule: JobSchedule{
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

		req, reqErr := http.NewRequest("POST", jobsURL, body)
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
		server := &http.Server{
			Addr: fmt.Sprintf("127.0.0.1:%d", *serverPort),
		}

		log.Printf("INFO: %d callbacks created", jobsCreated)
		log.Printf("INFO: Listening for callbacks on %s\n", server.Addr)
		log.Fatal(server.ListenAndServe())

	} else {
		log.Printf("INFO: No jobs were created, terminating.")
	}
}
