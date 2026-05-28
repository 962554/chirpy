// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/962554/chirpy/internal/auth"
	"github.com/962554/chirpy/internal/database"
	"github.com/google/uuid"
)

const (
	errTooLong = `{"error": "parameters is too long"}`
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
	Id      uuid.UUID `json:"id"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Body    string    `json:"body"`
	UserID  uuid.UUID `json:"user_id"`
}

func createChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeMessage(
			w,
			400,
			fmt.Appendf([]byte{}, errJSON, fmt.Sprintf("problem getting bearer token: %v", err)),
		)
		return
	}

	uid, err := auth.ValidateJWT(bearerToken, apiCfg.jwtSecret)
	if err != nil {
		writeMessage(w, 401, fmt.Appendf([]byte{}, errJSON, err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		writeMessage(w, 400, []byte(errChirp))
		return
	}
	if len(params.Body) > maxLength {
		writeMessage(w, 400, []byte(errTooLong))
		return
	}

	chirp, err := apiCfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: uid,
	})
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem creating chirp"))
		return
	}

	c := Chirp{
		Id:      chirp.ID,
		Created: chirp.CreatedAt,
		Updated: chirp.UpdatedAt,
		Body:    chirp.Body,
		UserID:  chirp.UserID,
	}

	dat, err := json.Marshal(c)
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem marshalling chirp to JSON"))
		return
	}
	writeMessage(w, 201, dat)
}

func allChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := apiCfg.dbQueries.AllChirps(r.Context())
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem fetching all chirps"))
		return
	}

	chirpsOut := []Chirp{}
	for _, chirp := range chirps {
		c := Chirp{
			Id:      chirp.ID,
			Created: chirp.CreatedAt,
			Updated: chirp.UpdatedAt,
			Body:    chirp.Body,
			UserID:  chirp.UserID,
		}
		chirpsOut = append(chirpsOut, c)
	}
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(chirpsOut)
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem marshalling chirps to JSON"))
		return
	}
	writeMessage(w, 200, dat)
}

func getChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "UUID parse error."))
		return
	}

	chirp, err := apiCfg.dbQueries.GetChirp(r.Context(), chirpID)

	if err == sql.ErrNoRows {
		writeMessage(w, 404, fmt.Appendf([]byte{}, errJSON, "chirp not found"))
		return
	}

	c := Chirp{
		Id:      chirp.ID,
		Created: chirp.CreatedAt,
		Updated: chirp.UpdatedAt,
		Body:    chirp.Body,
		UserID:  chirp.UserID,
	}

	dat, err := json.Marshal(c)
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem marshalling chirp to JSON"))
		return
	}
	writeMessage(w, 200, dat)
}

func deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeMessage(
			w,
			401,
			fmt.Appendf([]byte{}, errJSON, fmt.Sprintf("problem getting bearer token: %v", err)),
		)
		return
	}

	uid, err := auth.ValidateJWT(bearerToken, apiCfg.jwtSecret)
	if err != nil {
		writeMessage(w, 401, fmt.Appendf([]byte{}, errJSON, err.Error()))
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "UUID parse error."))
		return
	}

	chirp, err := apiCfg.dbQueries.GetChirp(r.Context(), chirpID)

	if err == sql.ErrNoRows {
		writeMessage(w, 404, fmt.Appendf([]byte{}, errJSON, "chirp not found"))
		return
	}

	if chirp.UserID != uid {
		writeMessage(w, 403, fmt.Appendf([]byte{}, errJSON, "chirp not owned by user"))
		return
	}

	err = apiCfg.dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		writeMessage(w, 404, fmt.Appendf([]byte{}, errJSON, "problem deleting chirp from db"))
		return
	}

	writeMessage(w, 204, []byte{})
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
