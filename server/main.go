package main

import (
	"log"
	"net/http"

	"github.com/marcoshack/schedula"
)

const (
	version  = "v0.1"
	jobsPath = "/jobs/"
)

func main() {
	log.Printf("Schedula Server %s", version)
	httpServer := &http.Server{
		Addr: "127.0.0.1:8080",
	}

	repository, err := schedula.NewRepository()
	scheduler, err := schedula.InitAndStartScheduler(repository, map[string]interface{}{})
	if err != nil {
		log.Fatalf("schedula: error initializing scheduler: %s", err)
	}

	http.Handle(jobsPath, &JobsHandler{scheduler: scheduler, Path: jobsPath})

	log.Printf("Listening on %s", httpServer.Addr)
	log.Fatal(httpServer.ListenAndServe())

	scheduler.Stop()
}
