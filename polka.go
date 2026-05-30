// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

const userEvent = "user.upgraded"

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem decoding parameters"))
		return
	}

	if params.Event != userEvent {
		writeMessage(w, 204, fmt.Appendf([]byte{}, errJSON, "wrong event type"))
		return
	}

	_, err = apiCfg.dbQueries.UpgradeUser(r.Context(), params.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeMessage(w, 404, fmt.Appendf([]byte{}, errJSON, "no such user"))
			return
		}
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem upgrading user to Chirpy Red"))
		return
	}
	writeMessage(w, 204, []byte{})
}
