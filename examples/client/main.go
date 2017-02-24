package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	Timeout         time.Duration
}

// CallbackURL creates a callback URL for the given `key`
func (c *ClientConfig) CallbackURL(key string) *url.URL {
	res, _ := url.Parse(fmt.Sprintf("%v%s", c.CallbackBaseURL, key))
	return res
}

var (
	nCallbacks    = flag.Int("n", 10, "`number` of callbacks to create")
	callbackPort  = flag.Int("p", 8088, "TCP `port` number to listen for HTTP callbacks")
	callbackAddr  = flag.String("b", "127.0.0.1", "IP `address` to listen for HTTP callbacks")
	serverBaseURL = flag.String("s", "http://localhost:8080/", "Schedula server base `URL`")
	callbackDelta = flag.Int("delta", 5, "delta in `seconds` from the current time to callbacks time")
	responseDelay = flag.Int("delay", 0, "delay in `milliseconds` to respond to callback request")
	timeout       = flag.Int("timeout", 30, "maximum number of `seconds` after callback time to wait for callbacks")
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
	timeout, _ := time.ParseDuration(fmt.Sprintf("%ds", *timeout))

	return &ClientConfig{
		TotalCallbacks:  *nCallbacks,
		CallbackAddr:    fmt.Sprintf("%s:%d", *callbackAddr, *callbackPort),
		CallbackBaseURL: callbackURL,
		JobsResourceURL: serverURL,
		CallbackDelta:   delta,
		CallbackTime:    time.Now().Add(delta),
		ResponseDelay:   delay,
		Timeout:         timeout,
	}
}

func startCallbackServer(conf *ClientConfig, callbacks chan int) {
	http.HandleFunc("/callback/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(conf.ResponseDelay)
		callbacks <- 1
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

	created := 0
	start := time.Now()
	for i := 1; i <= conf.TotalCallbacks; i++ {
		job := &Job{
			CallbackURL: conf.CallbackURL(fmt.Sprintf("?id=%d", i)).String(),
			Schedule: JobSchedule{
				Format: "timestamp",
				Value:  fmt.Sprintf("%v", conf.CallbackTime.Unix()),
			},
		}

		var body = new(bytes.Buffer)
		err := json.NewEncoder(body).Encode(job)
		if err != nil {
			log.Printf("ERROR: Unable to encode request body: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", conf.JobsResourceURL.String(), body)
		if err != nil {
			log.Printf("ERROR: Failed to create HTTP request: %v", err)
			continue
		}
		req.Header.Set("User-Agent", "schedula-client")
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			log.Printf("ERROR: Failed to send HTTP request: %v", err)
			continue
		}

		if res.StatusCode != http.StatusCreated {
			log.Printf("ERROR: Invalid response code, expected 201 Created but got %s", res.Status)
			continue
		}
		created++
	}

	if created == 0 {
		log.Printf("INFO: No jobs were created, terminating.")
		return
	}

	elapsed := time.Now().Sub(start).Seconds()
	rps := int(float64(created) / elapsed)
	log.Printf("INFO: %d callbacks created in %v seconds (~%d req/s)", created, elapsed, rps)

	timeout := make(chan bool, 1)
	go func() {
		timeoutTime := conf.CallbackTime.Add(conf.Timeout)
		log.Printf("INFO: Timeout set to %v", timeoutTime)
		time.Sleep(timeoutTime.Sub(time.Now()))
		timeout <- true
	}()

	callbacks := make(chan int, conf.TotalCallbacks)
	startCallbackServer(conf, callbacks)

	received := 0
ReceiveLoop:
	for {
		select {
		case <-callbacks:
			received++
			switch {
			case received%100 == 0 && received < created:
				log.Printf("INFO: %d/%d callbacks received", received, created)
			case received == created:
				log.Printf("INFO: %d/%d callbacks received. Terminating ...", received, created)
				break ReceiveLoop
			}
		case <-timeout:
			log.Printf("INFO: Timeout. %d callbacks received. Terminating ...", received)
			break ReceiveLoop
		}
	}
	log.Printf("INFO: Done.")
}
