// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

type email struct {
	Email string `json:"email"`
}

func createUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	email := email{}
	err := decoder.Decode(&email)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(errJSON, "problem decoding email")))
		return
	}

	user, err := apiCfg.dbQueries.CreateUser(r.Context(), email.Email)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(errJSON, "problem creating user")))
		return
	}

	u := User{
		Id:      user.ID,
		Created: user.CreatedAt,
		Updated: user.UpdatedAt,
		Email:   user.Email,
	}

	dat, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(errJSON, "problem marshalling user to JSON")))
		return
	}
	w.WriteHeader(201)
	w.Write(dat)
}
