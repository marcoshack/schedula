package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/marcoshack/schedula/Godeps/_workspace/src/github.com/gorilla/mux"
)

const version = "0.1"

var bindAddr = flag.String("b", "0.0.0.0", "IP `address` to bind")
var bindPort = flag.Int("p", 8080, "TCP `port` number to bind")
var nWorkers = flag.Int("w", 2, "number of `workers` to execute callback requests")

type config struct {
	BindAddr        string
	BindPort        int
	NumberOfWorkers int
}

func (c *config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.BindAddr, c.BindPort)
}

func readConfig() *config {
	flag.Parse()
	return &config{
		BindAddr:        *bindAddr,
		BindPort:        *bindPort,
		NumberOfWorkers: *nWorkers,
	}
}

func main() {
	log.Printf("Schedula Server v%s", version)
	config := readConfig()

	repository := initRepository()
	executor := initCallbackExecutor(repository)
	scheduler := initScheduler(repository, executor, config.NumberOfWorkers)

	jobs := &JobsHandler{repository: repository, path: "/jobs/"}
	router := mux.NewRouter()
	router.HandleFunc("/jobs/", jobs.List).Methods("GET")
	router.HandleFunc("/jobs/", jobs.Create).Methods("POST")
	router.HandleFunc("/jobs/{id}", jobs.Find).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobs.Delete).Methods("DELETE")

	log.Printf("Listening on %s", config.ServerAddr())
	log.Fatal(http.ListenAndServe(config.ServerAddr(), router))

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
