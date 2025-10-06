package api_test

import (
	"bytes"
	"encoding/json"
	"generic-shop-sample/internal/testutils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var baseAuthURL = "/api/auth"

func TestAuthHandler(t *testing.T) {
	app := testutils.RouterSetup(t.Context())

	tests := []struct {
		name        string
		code        int
		body        map[string]string
		tokenString string
		expectError bool
	}{
		{"login with valid user", http.StatusOK, map[string]string{"username": "admin_user", "password": "securePassword"}, "", false},
		{"login with valid user and invalid password", http.StatusUnauthorized, map[string]string{"username": "admin_user", "password": "notSetPassword123!@#$"}, "", true},
		{"login with invalid username", http.StatusNotFound, map[string]string{"username": "not-set-user1", "password": "securePassword"}, "", true},
		{"login with invalid username and password", http.StatusBadRequest, map[string]string{"username": "adminUser", "password": "secure Password"}, "", true},
	}

	for range 2 {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				reqJson, _ := json.Marshal(test.body)
				req, _ := http.NewRequest(http.MethodPost, baseAuthURL+"/login", bytes.NewBuffer(reqJson))
				app.ServeHTTP(w, req)

				if test.code != w.Code {
					t.Errorf(`expected status="%d", but got "%d"`, test.code, w.Code)
				}
				if test.expectError {
					return
				}
				var resJson map[string]string
				_ = json.NewDecoder(w.Body).Decode(&resJson)
				cookie := w.Header().Get("Set-Cookie")
				token := strings.TrimPrefix(resJson["token"], "Bearer ")
				if !strings.Contains(cookie, token) {
					t.Errorf(`unmatched response "%s" and cookie token "%s"`, resJson["token"], cookie)
				}

				if test.tokenString != "" && test.tokenString != token {
					t.Errorf(`failed to cache token "%s", got new token "%s"`, test.tokenString, token)
				}
				test.tokenString = token
			})
		}
	}
}
