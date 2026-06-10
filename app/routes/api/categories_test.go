package api_test

import (
	"bytes"
	"fmt"
	tu "generic-shop-sample/internal/testutils"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const baseCategoriesURL = "/api/categories/"

func TestCategoriesHandler(t *testing.T) {
	ctx := t.Context()

	app := tu.RouterSetup(ctx)
	adminToken := tu.LoginSetup(app, "admin_user", "securePassword")

	tests := []struct {
		name   string
		method string
		url    string
		code   int
		token  string
		body   io.Reader
	}{
		{
			"categories.create",
			http.MethodPost,
			baseCategoriesURL,
			http.StatusCreated,
			adminToken,
			bytes.NewBuffer([]byte(`{"tag": "some new tag"}`)),
		},
		{
			"categories.list",
			http.MethodGet,
			baseCategoriesURL,
			http.StatusOK,
			"",
			nil,
		},
		{
			"categories.delete",
			http.MethodDelete,
			fmt.Sprintf("%s1", baseCategoriesURL),
			http.StatusNoContent,
			adminToken,
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
		})
		time.Sleep(500 * time.Microsecond)
	}
}

const basePCURL = "/api/categories/pc/"

func TestPCHandler(t *testing.T) {
	ctx := t.Context()

	app := tu.RouterSetup(ctx)
	adminToken := tu.LoginSetup(app, "admin_user", "securePassword")

	session := database.GetSession()
	ps := queries.NewProductStore(session)
	if err := ps.Create(ctx, 1, &queries.CreateProductRequest{
		Name:        "new model",
		Price:       1233,
		Description: "some info",
		Details:     `{"yep": "yep"}`,
		IsAvailable: true,
	}); err != nil {
		t.Errorf("failed to create products, %s", err)
	}
	if _, err := session.Exec(ctx, "UPDATE products SET is_active = true"); err != nil {
		t.Errorf("failed to activate products, %s", err)
	}
	products, err := ps.List(ctx, 20, 1)
	if err != nil {
		t.Errorf("failed to get products list, %s", err)
	}

	cs := queries.NewCategoryStore(session)
	for _, tag := range []string{"a", "b", "c"} {
		if err := cs.Create(ctx, tag); err != nil {
			t.Errorf(`failed to create tag, %s, %s`, tag, err)
		}
	}

	// testing pcHandler.setTag
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%sset-tags/%s", basePCURL, products[0].ID), bytes.NewBuffer([]byte(`{"tags": ["a", "b", "c"]}`)))
	req.Header.Set("Authorization", adminToken)
	app.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf(`expected status="%d", but got "%d"`, http.StatusAccepted, w.Code)
	}

	// testing pcHandler.list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", basePCURL, products[0].ID), nil)
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(`expected status="%d", but got "%d"`, http.StatusOK, w.Code)
	}
}
