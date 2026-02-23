// Package Chirpy is a social network similar to Twitter.
package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	const (
		port        = ":8080"
		readTimeout = 5 * time.Second
	)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: readTimeout,
	}

	log.Printf("http server starting on port: %s", port)
	log.Fatal(server.ListenAndServe())
}
