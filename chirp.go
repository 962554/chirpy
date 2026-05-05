// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	errTooLong = `{"error": "Chirp is too long"}`
	errChirp   = `{"error": "Something went wrong"}`
	validChirp = `{"valid": true}`
	cleaned    = `{"cleaned_body": %q}`
	maxLength  = 140
	cleaner    = "****"
)

var profane = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

type Chirp struct {
	Body string `json:"body"`
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(errChirp))
		return
	}
	if len(chirp.Body) > maxLength {
		w.WriteHeader(400)
		w.Write([]byte(errTooLong))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(cleaned, clean(chirp.Body))))
}

func clean(in string) string {
	words := strings.Split(in, " ")
	cleaned := []string{}
	for _, word := range words {
		if profane[strings.ToLower(word)] == true {
			word = cleaner
		}
		cleaned = append(cleaned, word)
	}
	return strings.Join(cleaned, " ")
}
