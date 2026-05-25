// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

// This package handles authentication for Chirp users.

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	chirpyJWTIssuer = "chirpy-access"
	authHeaderKey   = "Authorization"
	bearerPrefix    = "Bearer "
)

// HashPassword hashes the provided password using argon2id.CreateHash
func HashPassword(password string) (string, error) {
	hashed, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashed, nil
}

// CheckPasswordHash compares the password that the user entered in the HTTP request with the password that is stored in the database.
func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    chirpyJWTIssuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(*jwt.Token) (any, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}
	userUUID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != chirpyJWTIssuer {
		return uuid.Nil, errors.New("invalid issuer")
	}
	uid, err := uuid.Parse(userUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return uid, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authz := headers.Get(authHeaderKey)
	if authz == "" {
		return "", errors.New("no auth header found")
	}
	token, found := strings.CutPrefix(authz, bearerPrefix)
	if !found {
		return "", errors.New("bearer prefix missing")
	}
	if token == "" {
		return "", errors.New("bearer token is empty")
	}
	return token, nil
}
