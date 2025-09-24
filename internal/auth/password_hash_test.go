package auth_test

import (
	"generic-shop-sample/internal/auth"
	"testing"
)

func TestPasswordHash(t *testing.T) {
	pwd := "SecurePassword"
	hash, err := auth.PasswordHash(pwd)
	if err != nil {
		t.Errorf("failed to generate hash from password, %s", err)
	}

	if !auth.ComparePassword(hash, pwd) {
		t.Errorf("failed to validate password")
	}
}
