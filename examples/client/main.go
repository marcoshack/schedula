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

var (
	numberOfCalbacks  = flag.Int("n", 10, "`number` of callbacks to create")
	serverPort        = flag.Int("p", 8088, "TCP `port` number to listen for HTTP callbacks")
	serverAddr        = flag.String("b", "127.0.0.1", "IP `address` to listen for HTTP callbacks")
	callbackTimeDelay = flag.Int("d", 5, "delay in `seconds` to create callbacks")
	serverBaseURL     = flag.String("s", "http://localhost:8080/", "Schedula server base `URL`")
)

func main() {
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
	start := time.Now()
	for i := 1; i <= *numberOfCalbacks; i++ {
		job := &Job{
			CallbackURL: fmt.Sprintf("http://%s:%d/callback/%d", *serverAddr, *serverPort, i),
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
		req.Header.Set("User-Agent", "schedula-client")
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
		jobsCreated++
	}

	elapsed := time.Now().Sub(start).Seconds()
	rps := int(float64(jobsCreated) / elapsed)

	if jobsCreated > 0 {
		log.Printf("INFO: %d callbacks created in %v seconds (~%d req/s)", jobsCreated, elapsed, rps)

		server := &http.Server{
			Addr: fmt.Sprintf("%s:%d", *serverAddr, *serverPort),
		}

		http.HandleFunc("/callback/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("INFO: Callback received %s", r.URL.Path)
		})

		log.Printf("INFO: Listening for callbacks on %s\n", server.Addr)
		log.Fatal(server.ListenAndServe())

	} else {
		log.Printf("INFO: No jobs were created, terminating.")
	}
}
