// -*- mode:go;mode:go-playground -*-
// Copyright © 2026 P, Rich
// License: MIT, see LICENSE for details

// This package handles authentication for Chirp users.

package auth

import (
	"log"

	"github.com/alexedwards/argon2id"
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
