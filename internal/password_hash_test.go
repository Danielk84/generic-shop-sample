package internal_test

import (
	"generic-shop-sample/internal"
	"testing"
)

func TestPasswordHash(t *testing.T) {
	pwd := "SecurePassword"
	hash, err := internal.PasswordHash(pwd)
	if err != nil {
		t.Errorf("failed to generate hash from password, %s", err)
	}

	if !internal.ComparePassword(hash, pwd) {
		t.Errorf("failed to validate password")
	}
}
