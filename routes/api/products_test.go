package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	tu "generic-shop-sample/internal/testutils"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const baseProductsURL = "/api/products/"

func TestProductsHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	app := tu.RouterSetup(ctx)
	adminToken := tu.LoginSetup(app, "admin_user", "securePassword")
	vendorToken := tu.LoginSetup(app, "vendor_user", "securePassword")

	session := db.NewSession()
	us := queries.NewUserStore(session)
	vendor, _ := us.Get(ctx, "vendor_user")

	if _, err := session.Exec(ctx, "TRUNCATE products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products, %s", err)
	}
	ps := queries.NewProductStore(session)
	for i := range 10 {
		if err := ps.Create(ctx, vendor.ID, &queries.CreateProductRequest{
			Name:        fmt.Sprintf("product - %d", i),
			Price:       10,
			Description: "some info",
			IsAvailable: true,
		}); err != nil {
			t.Errorf(`failed to create product "%d", %s`, i, err)
		}
	}

	if _, err := session.Exec(ctx, "UPDATE products SET is_active = true"); err != nil {
		t.Errorf("failed to activate products, %s", err)
	}

	products, err := ps.FullList(ctx, 0, 10, 1)
	if err != nil {
		t.Errorf("failed to get full products list, %s", err)
	}

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
			"productsHandler.list",
			http.MethodGet,
			baseProductsURL,
			nil,
			"",
			http.StatusOK,
			nil,
		},
		{
			"productsHandler.get - public",
			http.MethodGet,
			baseProductsURL + fmt.Sprintf("%s", products[0].ID),
			nil,
			"",
			http.StatusOK,
			nil,
		},
		{
			"productsHandler.fullList - admin",
			http.MethodGet,
			baseProductsURL + "user/",
			nil,
			adminToken,
			http.StatusOK,
			func(st *testing.T, w *httptest.ResponseRecorder) {
				var resJson []map[string]string
				_ = json.NewDecoder(w.Body).Decode(&resJson)
				if len(products) != len(resJson) {
					st.Errorf(`expected len="%d", but got "%d"`, len(products), len(resJson))
				}
			},
		},
		{
			"productsHandler.fullList - vendor",
			http.MethodGet,
			baseProductsURL + "user/",
			nil,
			vendorToken,
			http.StatusOK,
			nil,
		},
		{
			"productsHandler.get - user",
			http.MethodGet,
			baseProductsURL + fmt.Sprintf("user/%s", products[1].ID),
			nil,
			vendorToken,
			http.StatusOK,
			nil,
		},
		{
			"productsHandler.get - admin",
			http.MethodGet,
			baseProductsURL + fmt.Sprintf("user/%s", products[2].ID),
			nil,
			adminToken,
			http.StatusOK,
			nil,
		},
		{
			"productsHandler.update",
			http.MethodPut,
			baseProductsURL + "user/",
			bytes.NewBuffer(fmt.Appendf([]byte(""), `{"id": "%s", "name": "new name", "price": 123, "description": "new info", "details": "{\"some-feild\": \"some info\"}", "is_available": true}`, products[6].ID)),
			vendorToken,
			http.StatusAccepted,
			nil,
		},
		{
			"productsHandler.setAvailable",
			http.MethodPut,
			baseProductsURL + fmt.Sprintf("user/set-available/%s", products[3].ID),
			bytes.NewBuffer([]byte(`{"accepted": true}`)),
			vendorToken,
			http.StatusAccepted,
			nil,
		},
		{
			"productHandler.setActive",
			http.MethodPut,
			baseProductsURL + fmt.Sprintf("user/set-active/%s", products[4].ID),
			bytes.NewBuffer([]byte(`{"accepted": true}`)),
			adminToken,
			http.StatusAccepted,
			nil,
		},
		{
			"productsHandler.delete",
			http.MethodDelete,
			baseProductsURL + fmt.Sprintf("user/%s", products[5].ID),
			nil,
			vendorToken,
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
	}
}

const baseCategoriesURL = "/api/categories/"

func TestCategoriesHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

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
			baseCategoriesURL + "user/",
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
			baseCategoriesURL + "user/1",
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

const basePCURL = "/api/pc/"

func TestPCHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	app := tu.RouterSetup(ctx)
	adminToken := tu.LoginSetup(app, "admin_user", "securePassword")

	session := db.NewSession()
	ps := queries.NewProductStore(session)
	products, err := ps.List(ctx, 20, 1)
	if err != nil {
		t.Errorf("failed to get products list, %s", err)
	}

	cs := queries.NewCategoryStore(session)
	for _, tag := range []string{"a", "b", "c"} {
		if err := cs.Create(ctx, tag); err != nil {
			t.Errorf("failed to create tag, %s", tag)
		}
	}

	// testing pcHandler.setTag
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, basePCURL+"set-tags/"+products[0].ID, bytes.NewBuffer([]byte(`{"tags": ["a", "b", "c"]}`)))
	req.Header.Set("Authorization", adminToken)
	app.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf(`expected status="%d", but got "%d"`, http.StatusAccepted, w.Code)
	}

	// testing pcHandler.list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, basePCURL+products[0].ID, nil)
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(`expected status="%d", but got "%d"`, http.StatusOK, w.Code)
	}
}
