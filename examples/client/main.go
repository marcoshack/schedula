package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
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

// ClientConfig ...
type ClientConfig struct {
	TotalCallbacks  int
	CallbackBaseURL *url.URL
	CallbackAddr    string
	JobsResourceURL *url.URL
	CallbackDelta   time.Duration
	CallbackTime    time.Time
	ResponseDelay   time.Duration
}

// CallbackURL creates a callback URL for the given `key`
func (c *ClientConfig) CallbackURL(key string) *url.URL {
	res, _ := url.Parse(fmt.Sprintf("%v%s", c.CallbackBaseURL, key))
	return res
}

type callbackCounter struct {
	sync.RWMutex
	count int
}

var (
	nCallbacks    = flag.Int("n", 10, "`number` of callbacks to create")
	callbackPort  = flag.Int("p", 8088, "TCP `port` number to listen for HTTP callbacks")
	callbackAddr  = flag.String("b", "127.0.0.1", "IP `address` to listen for HTTP callbacks")
	serverBaseURL = flag.String("s", "http://localhost:8080/", "Schedula server base `URL`")
	callbackDelta = flag.Int("delta", 5, "delta in `seconds` from the current time to create callbacks")
	responseDelay = flag.Int("delay", 0, "delay in `milliseconds` to respond to callback request")
)

func loadConfig() *ClientConfig {
	flag.Parse()
	serverURL, err := url.Parse(fmt.Sprintf("%sjobs/", *serverBaseURL))
	if err != nil {
		log.Fatalf("ERROR: invalid server base URL: %v", err)
	}
	callbackURL, err := url.Parse(fmt.Sprintf("http://%s:%d/callback/", *callbackAddr, *callbackPort))
	if err != nil {
		log.Fatalf("ERROR: invalid callback URL: %v", err)
	}
	delta, _ := time.ParseDuration(fmt.Sprintf("%ds", *callbackDelta))
	delay, _ := time.ParseDuration(fmt.Sprintf("%dms", *responseDelay))

	return &ClientConfig{
		TotalCallbacks:  *nCallbacks,
		CallbackAddr:    fmt.Sprintf("%s:%d", *callbackAddr, *callbackPort),
		CallbackBaseURL: callbackURL,
		JobsResourceURL: serverURL,
		CallbackDelta:   delta,
		CallbackTime:    time.Now().Add(delta),
		ResponseDelay:   delay,
	}
}

func startCallbackServer(conf *ClientConfig, jobsCreated int, done chan int) {
	counter := callbackCounter{}
	http.HandleFunc("/callback/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(conf.ResponseDelay)
		counter.Lock()
		counter.count++
		if counter.count%100 == 0 || counter.count == jobsCreated {
			log.Printf("INFO: %d/%d callback(s) received", counter.count, jobsCreated)
		}
		if counter.count == jobsCreated {
			log.Printf("INFO: all callbacks received. terminating")
			done <- 1
		}
		counter.Unlock()
	})
	server := &http.Server{Addr: conf.CallbackAddr}
	log.Printf("INFO: Listening for callbacks on %s\n", server.Addr)
	go server.ListenAndServe()
}

func checkServer(conf *ClientConfig, client *http.Client) {
	_, err := client.Head(conf.JobsResourceURL.String())
	if err != nil {
		log.Fatalf("ERROR: schedula server is unavailable: %v", err)
	}
}

func main() {
	conf := loadConfig()
	client := &http.Client{}
	checkServer(conf, client)

	jobsCreated := 0
	start := time.Now()
	for i := 1; i <= conf.TotalCallbacks; i++ {
		job := &Job{
			CallbackURL: conf.CallbackURL(fmt.Sprintf("%d", i)).String(),
			Schedule: JobSchedule{
				Format: "timestamp",
				Value:  fmt.Sprintf("%v", conf.CallbackTime.Unix()),
			},
		}

		var body = new(bytes.Buffer)
		err := json.NewEncoder(body).Encode(job)
		if err != nil {
			log.Printf("ERROR: unable to encode request body: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", conf.JobsResourceURL.String(), body)
		if err != nil {
			log.Printf("ERROR: failed to create HTTP request: %v", err)
			continue
		}
		req.Header.Set("User-Agent", "schedula-client")
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			log.Printf("ERROR: failed to send HTTP request: %v", err)
			continue
		}

		if res.StatusCode != http.StatusCreated {
			log.Printf("ERROR: invalid response code, expected 201 Created but got %s", res.Status)
			continue
		}
		jobsCreated++
	}

	if jobsCreated == 0 {
		log.Printf("INFO: No jobs were created, terminating.")
		return
	}

	elapsed := time.Now().Sub(start).Seconds()
	rps := int(float64(jobsCreated) / elapsed)
	log.Printf("INFO: %d callbacks created in %v seconds (~%d req/s)", jobsCreated, elapsed, rps)

	done := make(chan int)
	startCallbackServer(conf, jobsCreated, done)
	<-done
}
