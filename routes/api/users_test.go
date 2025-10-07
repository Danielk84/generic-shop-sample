package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	tu "generic-shop-sample/internal/testutils"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

var baseUserURL = "/api/users/"

func TestUsersHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	app := tu.RouterSetup(ctx)
	us := queries.NewUserStore(db.NewSession())
	adminToken := tu.LoginSetup(app, "admin_user", "securePassword")
	admin2Token := tu.LoginSetup(app, "admin_user2", "securePassword")
	customerToken := tu.LoginSetup(app, "customer_user", "securePassword")

	blockUser, _ := us.Get(ctx, "block_user")

	tests := []struct {
		name   string
		method string
		url    string
		body   io.Reader
		token  string
		code   int
		after  func(st *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			"usersHandler.createUserByAdmin",
			http.MethodPost,
			baseUserURL,
			bytes.NewBuffer([]byte(`{"username": "new-user-by-admin", "password": "securePassword"}`)),
			adminToken,
			http.StatusCreated,
			func(st *testing.T, w *httptest.ResponseRecorder) {
				if isExists := us.IsUsernameExists(ctx, "new-user-by-admin"); !isExists {
					st.Errorf("failed to create new user by admin")
				}
			},
		},
		{
			"usersHandler.list",
			http.MethodGet,
			baseUserURL,
			nil,
			adminToken,
			http.StatusOK,
			func(st *testing.T, w *httptest.ResponseRecorder) {
				var resJson []map[string]string
				_ = json.NewDecoder(w.Body).Decode(&resJson)
				if resJson[0]["password"] != "" {
					st.Errorf(`bad serializer`)
				}
			},
		},
		{
			"usersHandler.get",
			http.MethodGet,
			baseUserURL + "admin_user",
			nil,
			adminToken,
			http.StatusOK,
			func(st *testing.T, w *httptest.ResponseRecorder) {
				var resJson map[string]string
				_ = json.NewDecoder(w.Body).Decode(&resJson)
				if got := resJson["username"]; got != "admin_user" {
					t.Errorf(`unexpected user "%s"`, got)
				}
			},
		},
		{
			"userHandler.updateUserPermission",
			http.MethodPut,
			baseUserURL + strconv.Itoa(int(blockUser.ID)),
			bytes.NewBuffer([]byte(`{"permission_type": 3, "is_active": false}`)),
			adminToken,
			http.StatusAccepted,
			nil,
		},
		{
			"userHandler.delete",
			http.MethodDelete,
			baseUserURL,
			nil,
			admin2Token,
			http.StatusNoContent,
			nil,
		},
		{
			"userHandler.delete - redelete user for check auth",
			http.MethodDelete,
			baseUserURL,
			nil,
			admin2Token,
			http.StatusNotFound,
			nil,
		},
		{
			"usersHandler.setEmail",
			http.MethodPut,
			baseUserURL + "set-email",
			bytes.NewBuffer([]byte(`{"email": "customer@bib.com"}`)),
			customerToken,
			http.StatusAccepted,
			nil,
		},
		{
			"usersHandler.setEmail - check uniqueness",
			http.MethodPut,
			baseUserURL + "set-email",
			bytes.NewBuffer([]byte(`{"email": "customer@bib.com"}`)),
			adminToken,
			http.StatusBadRequest,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(st *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.method, test.url, test.body)
			req.Header.Set("Authorization", test.token)
			app.ServeHTTP(w, req)
			if w.Code != test.code {
				st.Errorf(`expected status="%d", but got "%d"`, test.code, w.Code)
			}
			if test.after != nil {
				test.after(st, w)
			}
		})
	}
}
