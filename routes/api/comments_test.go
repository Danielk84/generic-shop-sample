package api_test

import (
	"bytes"
	"context"
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	tu "generic-shop-sample/internal/testutils"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

const baseCommentsURL = "/api/comments/"

func TestCommentsHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	app := tu.RouterSetup(ctx)
	session := db.NewSession()
	if _, err := session.Exec(ctx, "TRUNCATE comments RESTART IDENTITY"); err != nil {
		t.Errorf("failed to truncate comments, %s", err)
	}

	us := queries.NewUserStore(session)
	customer, err := us.Get(ctx, "customer_user")
	if err != nil {
		t.Errorf("failed to get customer user, %s", err)
	}
	customerToken := tu.LoginSetup(app, "customer_user", "securePassword")
	adminToken := tu.LoginSetup(app, "admin_user", "securePassword")

	cs := queries.NewCommentStore(session)
	referrer := url.QueryEscape("/")
	if err = cs.Create(ctx, &queries.CommentRequest{
		Username: customer.Username,
		Parent:   "",
		Referrer: referrer,
		Body:     "yeppi",
	}); err != nil {
		t.Errorf("failed to create parent comment, %s", err)
	}
	if _, err := session.Exec(ctx, "UPDATE comments SET is_active = true"); err != nil {
		t.Errorf("failed to activate comments, %s", err)
	}
	parents, err := cs.List(ctx, "", referrer, 20, 1)
	if err != nil {
		t.Errorf("failed to get comments list, %s", err)
	}
	parent := parents[0]

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
			"commentsHandler.create",
			http.MethodPost,
			baseCommentsURL + "user/",
			bytes.NewBuffer(fmt.Appendf([]byte(""), `{"username": "customer_user", "parent": "%s", "referrer": "%s", "body": "some body"}`, parent.ID, referrer)),
			customerToken,
			http.StatusCreated,
			func(st *testing.T, w *httptest.ResponseRecorder) {
				if _, err := session.Exec(ctx, "UPDATE comments SET is_active = true"); err != nil {
					st.Errorf("failed to activate comments, %s", err)
				}
			},
		},
		{
			"commentsHandler.get",
			http.MethodGet,
			baseCommentsURL + fmt.Sprintf("user/%s", parent.ID),
			nil,
			customerToken,
			http.StatusOK,
			nil,
		},
		{
			"commentsHandler.list",
			http.MethodGet,
			baseCommentsURL + fmt.Sprintf("?parent=%s&referrer=%s", parent.ID, referrer),
			nil,
			"",
			http.StatusOK,
			nil,
		},
		{
			"commentsHandler.fullList",
			http.MethodGet,
			baseCommentsURL + "user/",
			nil,
			customerToken,
			http.StatusOK,
			nil,
		},
		{
			"commentsHandler.setActive",
			http.MethodPut,
			baseCommentsURL + fmt.Sprintf("user/set-active/%s", parent.ID),
			bytes.NewBuffer([]byte(`{"accepted": true}`)),
			adminToken,
			http.StatusAccepted,
			nil,
		},
		{
			"commentsHandler.delete",
			http.MethodDelete,
			baseCommentsURL + fmt.Sprintf("user/%s", parent.ID),
			nil,
			customerToken,
			http.StatusNoContent,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(st *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.method, test.url, test.body)
			if test.token != "" {
				req.Header.Set("Authorization", test.token)
			}
			app.ServeHTTP(w, req)
			if w.Code != test.code {
				st.Errorf(`expected status="%d", but got "%d"`, test.code, w.Code)
			}
			if test.after != nil {
				test.after(st, w)
			}
		})
		time.Sleep(500 * time.Millisecond)
	}
}
