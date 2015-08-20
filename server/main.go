package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/marcoshack/schedula"
)

const (
	version  = "v0.1"
	jobsPath = "/jobs/"
)

func main() {
	log.Printf("Schedula Server %s", version)

	serverAddr := flag.String("b", "127.0.0.1:8080", "HTTP bind address")
	numberOfWorkers := flag.Int("w", 50, "number of workers to execute callback requests")
	flag.Parse()

	repository, err := schedula.NewRepository()
	scheduler, err := schedula.InitAndStartScheduler(repository, schedula.SchedulerConfig{
		NumberOfWorkers: *numberOfWorkers,
	})
	if err != nil {
		log.Fatalf("schedula: error initializing scheduler: %s", err)
	}

	jobsHandler := &JobsHandler{scheduler: scheduler, Path: jobsPath}

	router := mux.NewRouter()
	router.HandleFunc("/jobs/", jobsHandler.List).Methods("GET")
	router.HandleFunc("/jobs/", jobsHandler.Create).Methods("POST")
	router.HandleFunc("/jobs/{id}", jobsHandler.Find).Methods("GET")

	log.Printf("Listening on %s", *serverAddr)
	log.Fatal(http.ListenAndServe(*serverAddr, router))

	scheduler.Stop()
}
