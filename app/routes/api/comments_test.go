package api_test

import (
	"bytes"
	"fmt"
	"generic-shop-sample/internal/logger"
	tu "generic-shop-sample/internal/testutils"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const baseCommentsURL = "/api/comments/"

func TestCommentsHandler(t *testing.T) {
	ctx := t.Context()

	app := tu.RouterSetup(ctx)
	session := database.GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE comments, products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate comments, %s", err)
	}

	log := logger.GetLogger()
	us := queries.NewUserStore(session, log)
	customer, err := us.Get(ctx, "customer_user")
	if err != nil {
		t.Errorf("failed to get customer user, %s", err)
	}
	customerToken := tu.LoginSetup(app, "customer_user", "securePassword")
	adminToken := tu.LoginSetup(app, "admin_user", "securePassword")

	ps := queries.NewProductStore(session)
	if err := ps.Create(ctx, 1, &queries.CreateProductRequest{
		Name:        "new product for comments",
		Price:       1001,
		Description: "lalala",
		Details:     "{}",
		IsAvailable: true,
	}); err != nil {
		t.Errorf("failed to create product, %s", err)
	}
	products, err := ps.FullList(ctx, 0, 20, 1)
	if err != nil {
		t.Errorf("failed to get products full list, %s", err)
	}
	product := products[0]

	cs := queries.NewCommentStore(session)
	referrer := product.ID
	if err = cs.Create(ctx, customer.Username, &queries.CommentRequest{
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
			baseCommentsURL,
			bytes.NewBuffer(fmt.Appendf([]byte(""), `{"referrer": "%s", "body": "some body"}`, referrer)),
			customerToken,
			http.StatusCreated,
			nil,
		},
		{
			"commentsHandler.create - subCommant",
			http.MethodPost,
			baseCommentsURL,
			bytes.NewBuffer(fmt.Appendf([]byte(""), `{"parent": "%s", "referrer": "%s", "body": "some body"}`, parent.ID, referrer)),
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
			fmt.Sprintf("%soverview/%s", baseCommentsURL, parent.ID),
			nil,
			customerToken,
			http.StatusOK,
			nil,
		},
		{
			"commentsHandler.list",
			http.MethodGet,
			fmt.Sprintf("%s?parent=%s&referrer=%s", baseCommentsURL, parent.ID, referrer),
			nil,
			"",
			http.StatusOK,
			nil,
		},
		{
			"commentsHandler.fullList",
			http.MethodGet,
			fmt.Sprintf("%sfull", baseCommentsURL),
			nil,
			customerToken,
			http.StatusOK,
			nil,
		},
		{
			"commentsHandler.setActive",
			http.MethodPut,
			fmt.Sprintf("%sset-active/%s", baseCommentsURL, parent.ID),
			bytes.NewBuffer([]byte(`{"accepted": true}`)),
			adminToken,
			http.StatusAccepted,
			nil,
		},
		{
			"commentsHandler.delete",
			http.MethodDelete,
			fmt.Sprintf("%s%s", baseCommentsURL, parent.ID),
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
