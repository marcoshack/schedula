package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/marcoshack/schedula/Godeps/_workspace/src/github.com/gorilla/mux"
)

const (
	version = "v0.1"
)

func main() {
	log.Printf("Schedula Server %s", version)

	serverAddr := flag.String("b", "127.0.0.1:8080", "HTTP bind `address`")
	nWorkers := flag.Int("w", 2, "`number` of workers to execute callback requests")
	flag.Parse()

	repository := initRepository()
	scheduler := initScheduler(repository, *nWorkers)
	jobsHandler := &JobsHandler{repository: repository, path: "/jobs/"}

	router := mux.NewRouter()
	router.HandleFunc("/jobs/", jobsHandler.List).Methods("GET")
	router.HandleFunc("/jobs/", jobsHandler.Create).Methods("POST")
	router.HandleFunc("/jobs/{id}", jobsHandler.Find).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobsHandler.Delete).Methods("DELETE")

	log.Printf("Listening on %s", *serverAddr)
	log.Fatal(http.ListenAndServe(*serverAddr, router))

	scheduler.Stop()
}

func initRepository() Repository {
	repository, repoErr := NewRepository()
	if repoErr != nil {
		log.Fatalf("schedula: error initializing repository: %v", repoErr)
	}
	return repository
}

func initScheduler(repository Repository, nWorkers int) Scheduler {
	scheduler, err := InitAndStartScheduler(repository, SchedulerConfig{NumberOfWorkers: nWorkers})
	if err != nil {
		log.Fatalf("schedula: error initializing scheduler: %v", err)
	}
	return scheduler
}
