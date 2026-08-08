package main

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashCanBeVerified(t *testing.T) {
	password := "admin-password-123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		t.Fatalf("generated password hash cannot be verified: %v", err)
	}
}
