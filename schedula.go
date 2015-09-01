package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/marcoshack/schedula/callback"
	"github.com/marcoshack/schedula/handler"
	"github.com/marcoshack/schedula/repository"
	"github.com/marcoshack/schedula/scheduler"
)

const version = "0.1"

var bindAddr = flag.String("b", "0.0.0.0", "IP `address` to bind")
var bindPort = flag.Int("p", 8080, "TCP `port` number to bind")
var nWorkers = flag.Int("w", 2, "number of `workers` to execute callback requests")
var repoType = flag.String("repo-type", "in-memory", "Repository `type` (available: in-memory)")
var schedType = flag.String("sched-type", "ticker", "Scheduler `type` (available: ticker)")

type config struct {
	BindAddr        string
	BindPort        int
	NumberOfWorkers int
	RepositoryType  string
	SchedulerType   string
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
		RepositoryType:  *repoType,
		SchedulerType:   *schedType,
	}
}

func main() {
	log.Printf("Schedula Server v%s", version)
	config := readConfig()

	repository := initRepository(config.RepositoryType)
	executor := initCallbackExecutor(repository)
	scheduler := initScheduler(config.SchedulerType, repository, executor, config.NumberOfWorkers)

	jobs := handler.NewJobsHandler("/jobs", repository)
	router := mux.NewRouter()
	router.HandleFunc("/jobs/", jobs.List).Methods("GET")
	router.HandleFunc("/jobs/", jobs.Create).Methods("POST")
	router.HandleFunc("/jobs/{id}", jobs.Find).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobs.Delete).Methods("DELETE")

	log.Printf("Listening on %s", config.ServerAddr())
	log.Fatal(http.ListenAndServe(config.ServerAddr(), router))

	scheduler.Stop()
}

func initRepository(repoType string) repository.Repository {
	repository, err := repository.New(repoType)
	if err != nil {
		log.Fatalf("schedula: error initializing repository: %v", err)
	}
	return repository
}

func initCallbackExecutor(repository repository.Repository) callback.Executor {
	executor, err := callback.NewExecutor()
	if err != nil {
		log.Fatalf("schedula: error initializing callback executor: %v", err)
	}
	return executor
}

func initScheduler(schedulerType string, r repository.Repository, e callback.Executor, nWorkers int) scheduler.Scheduler {
	scheduler, err := scheduler.StartNew(schedulerType, r, e, scheduler.Config{NumberOfWorkers: nWorkers})
	if err != nil {
		log.Fatalf("schedula: error initializing scheduler: %v", err)
	}
	return scheduler
}
