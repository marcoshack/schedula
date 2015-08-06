package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/marcoshack/schedula"
)

const (
	version = "v0.1"
)

func main() {
	httpServer := &http.Server{
		Addr: "127.0.0.1:8080",
	}

	scheduler, err := schedula.InitScheduler("in-memory")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/schedules/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprint(w, "{\"message\":\"Hello!\"}")
		case "POST":
			job := &schedula.Job{CallbackURL: "http://example.com/callback"}
			id, err := scheduler.Schedule(job)
			if err != nil {
				fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Add("Location", fmt.Sprintf("/schedules/%s", id))
				w.WriteHeader(http.StatusCreated)
			}
		}
	})

	log.Printf("Schedula Server %s started", version)
	log.Printf("Listening on %s", httpServer.Addr)
	log.Printf("Scheduler type '%s'", scheduler.Type())
	httpServer.ListenAndServe()
}
