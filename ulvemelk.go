package main

import (
	"log"
	"net/http"
)

func main() {
	srv := newServer()
	srv.setupRouting()
	if err := http.ListenAndServe(":4321", srv.mux); err != nil {
		log.Fatal(err)
	}
}
