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
		writeMessage(w, 400, []byte(errChirp))
		return
	}
	if len(chirp.Body) > maxLength {
		writeMessage(w, 400, []byte(errTooLong))
		return
	}
	writeMessage(w, http.StatusOK, fmt.Appendf([]byte{}, cleaned, clean(chirp.Body)))
}

func clean(in string) string {
	words := strings.Split(in, " ")
	cleaned := []string{}
	for _, word := range words {
		if profane[strings.ToLower(word)] {
			word = cleaner
		}
		cleaned = append(cleaned, word)
	}
	return strings.Join(cleaned, " ")
}
