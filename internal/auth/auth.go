// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

// This package handles authentication for Chirp users.

package auth

import (
	"encoding/json"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// HashPassword hashes the provided password using argon2id.CreateHash
func HashPassword(password string) (string, error) {
	hashed, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	log.Println("hash: original:", password, hashed)
	return hashed, nil
}

// CheckPasswordHash compares the password that the user entered in the HTTP request with the password that is stored in the database.
func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	uid, err := json.Marshal(userID)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  &jwt.NumericDate{time.Now().UTC()},
		ExpiresAt: &jwt.NumericDate{time.Now().Add(expiresIn)},
		Subject:   string(uid),
	})

	return token.SignedString(tokenSecret)
}

// func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
// jwt.ParseWithClaims(tokenString, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption)

// }
