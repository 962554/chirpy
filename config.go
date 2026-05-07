// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/962554/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) showHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writeMessage(w, http.StatusOK, fmt.Appendf([]byte{}, adminTemplate, cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		writeMessage(w, 403, []byte{})
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.dbQueries.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("error deleting users from db: %s", err)
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writeMessage(w, http.StatusOK, []byte{})
}
