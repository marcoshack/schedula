package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/marcoshack/schedula/Godeps/_workspace/src/github.com/gorilla/mux"
)

const (
	version = "v0.1"
)

func main() {
	log.Printf("Schedula Server %s", version)
	bindAddr := flag.String("b", "0.0.0.0", "IP `address` to bind")
	bindPort := flag.Int("p", 8080, "TCP `port` number to bind")
	nWorkers := flag.Int("w", 2, "number of `workers` to execute callback requests")
	flag.Parse()

	repository := initRepository()
	executor := initCallbackExecutor(repository)
	scheduler := initScheduler(repository, executor, *nWorkers)

	jobs := &JobsHandler{repository: repository, path: "/jobs/"}
	router := mux.NewRouter()
	router.HandleFunc("/jobs/", jobs.List).Methods("GET")
	router.HandleFunc("/jobs/", jobs.Create).Methods("POST")
	router.HandleFunc("/jobs/{id}", jobs.Find).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobs.Delete).Methods("DELETE")

	serverAddr := fmt.Sprintf("%s:%d", *bindAddr, *bindPort)
	log.Printf("Listening on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, router))

	scheduler.Stop()
}

func initRepository() Repository {
	repository, err := NewRepository()
	if err != nil {
		log.Fatalf("schedula: error initializing repository: %v", err)
	}
	return repository
}

func initCallbackExecutor(repository Repository) CallbackExecutor {
	executor, err := NewCallbackExecutor()
	if err != nil {
		log.Fatalf("schedula: error initializing callback executor: %v", err)
	}
	return executor
}

func initScheduler(repository Repository, executor CallbackExecutor, nWorkers int) Scheduler {
	scheduler, err := InitAndStartScheduler(repository, executor, SchedulerConfig{NumberOfWorkers: nWorkers})
	if err != nil {
		log.Fatalf("schedula: error initializing scheduler: %v", err)
	}
	return scheduler
}
