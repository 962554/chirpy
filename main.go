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

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	})
	log.Printf("http server starting on port: %s", port)
	log.Fatal(server.ListenAndServe())
}
