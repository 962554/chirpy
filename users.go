// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

package main

import (
	"database/sql"
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
	user, err := apiCfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	})
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

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	expiresIn := time.Hour

	jwtToken, err := auth.MakeJWT(user.ID, apiCfg.jwtSecret, expiresIn)
	if err != nil {
		writeMessage(
			w,
			400,
			fmt.Appendf([]byte{}, errJSON, fmt.Sprintf("problem creating JWT jwtToken: %v", err)),
		)
		return
	}

	refreshToken := auth.MakeRefreshToken()
	err = apiCfg.dbQueries.AddToken(r.Context(), database.AddTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem adding refreshToken to db"))
		return
	}

	resp := response{
		User: User{
			Id:      user.ID,
			Created: user.CreatedAt,
			Updated: user.UpdatedAt,
			Email:   user.Email,
		},
		Token:        jwtToken,
		RefreshToken: refreshToken,
	}

	dat, err := json.Marshal(resp)
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem marshalling user to JSON"))
		return
	}
	writeMessage(w, 200, dat)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type response struct {
		User
	}

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

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem decoding parameters"))
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		writeMessage(w, 400, fmt.Appendf([]byte{}, errJSON, "problem hashing password"))
	}

	u, err := apiCfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		HashedPassword: hash,
		Email:          params.Email,
		ID:             uid,
	})
	if err != nil {
		writeMessage(w, 401, fmt.Appendf([]byte{}, errJSON, "problem updating user"))
	}
	resp := response{
		User: User{
			Id:      u.ID,
			Created: u.CreatedAt,
			Updated: u.UpdatedAt,
			Email:   u.Email,
		},
	}
	dat, err := json.Marshal(resp)
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem marshalling user to JSON"))
		return
	}
	writeMessage(w, 200, dat)
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeMessage(
			w,
			400,
			fmt.Appendf([]byte{}, errJSON, fmt.Sprintf("problem getting bearer token: %v", err)),
		)
		return
	}

	refreshToken, err := apiCfg.dbQueries.GetToken(r.Context(), bearerToken)
	// if err != nil {
	// 	writeMessage(
	// 		w,
	// 		400,
	// 		fmt.Appendf([]byte{}, errJSON, fmt.Sprintf("problem getting refresh token from db: %v", err)),
	// 	)
	// 	return
	// }
	if err == sql.ErrNoRows || refreshToken.ExpiresAt.Before(time.Now()) || refreshToken.RevokedAt.Valid {
		writeMessage(
			w,
			401,
			fmt.Appendf([]byte{}, errJSON, "problem with refresh token"),
		)
		return
	}
	var jwtToken string
	if refreshToken.Token == bearerToken {
		jwtToken, err = auth.MakeJWT(refreshToken.UserID, apiCfg.jwtSecret, time.Hour)
		if err != nil {
			writeMessage(
				w,
				400,
				fmt.Appendf([]byte{}, errJSON, fmt.Sprintf("problem creating JWT jwtToken: %v", err)),
			)
			return
		}
	}
	dat, err := json.Marshal(struct {
		Token string `json:"token"`
	}{
		Token: jwtToken,
	})
	if err != nil {
		writeMessage(w, 500, fmt.Appendf([]byte{}, errJSON, "problem marshalling jwtToken to JSON"))
		return
	}
	writeMessage(w, 200, dat)
}

func revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeMessage(
			w,
			400,
			fmt.Appendf([]byte{}, errJSON, fmt.Sprintf("problem getting bearer token: %v", err)),
		)
		return
	}

	err = apiCfg.dbQueries.RevokeToken(r.Context(), bearerToken)
	if err != nil {
		writeMessage(
			w,
			400,
			fmt.Appendf([]byte{}, errJSON, fmt.Sprintf("problem getting bearer token: %v", err)),
		)
		return
	}
	writeMessage(
		w,
		204,
		fmt.Append([]byte{}, ""),
	)
	return
}
