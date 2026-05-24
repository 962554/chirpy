// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/962554/chirpy/internal/auth"
	"github.com/962554/chirpy/internal/database"
	"github.com/google/uuid"
)

const (
	errJSON = `{"error": %q}`
)

type User struct {
	Id      uuid.UUID `json:"id"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Email   string    `json:"email"`
}

var chirpUser = User{}

func createUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem decoding parameters"))
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem hashing password"))
	}
	user, err := apiCfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hash})
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem creating user"))
		return
	}

	chirpUser = User{
		Id:      user.ID,
		Created: user.CreatedAt,
		Updated: user.UpdatedAt,
		Email:   user.Email,
	}

	dat, err := json.Marshal(chirpUser)
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem marshalling user to JSON"))
		return
	}
	writeMessage(w, 201, dat)
}

func loginUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem decoding parameters"))
		return
	}

	user, err := apiCfg.dbQueries.GetUser(r.Context(), params.Email)
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem getting user from db"))
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		writeMessage(w, 401, fmt.Appendf([]byte{}, errJSON, "Incorrect email or password"))
		return
	}
	chirpUser = User{
		Id:      user.ID,
		Created: user.CreatedAt,
		Updated: user.UpdatedAt,
		Email:   user.Email,
	}

	dat, err := json.Marshal(chirpUser)
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem marshalling user to JSON"))
		return
	}
	writeMessage(w, 200, dat)
}
