package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/logger"
	tu "generic-shop-sample/internal/testutils"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const baseUserURL = "/api/users/"

func TestUsersHandler(t *testing.T) {
	ctx := t.Context()

	app := tu.RouterSetup(ctx)
	log := logger.GetLogger()
	us := queries.NewUserStore(database.GetSession(), log)
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
			fmt.Sprintf("%sadmin_user", baseUserURL),
			nil,
			"",
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
			fmt.Sprintf("%sset-email", baseUserURL),
			bytes.NewBuffer([]byte(`{"email": "customer@bib.com"}`)),
			customerToken,
			http.StatusAccepted,
			nil,
		},
		{
			"usersHandler.setEmail - check uniqueness",
			http.MethodPut,
			fmt.Sprintf("%sset-email", baseUserURL),
			bytes.NewBuffer([]byte(`{"email": "customer@bib.com"}`)),
			adminToken,
			http.StatusBadRequest,
			nil,
		},
		{
			"usersHandler.setPhoneNumber",
			http.MethodPut,
			fmt.Sprintf("%sset-phone-number", baseUserURL),
			bytes.NewBuffer([]byte(`{"phone_number": "09999999999"}`)),
			adminToken,
			http.StatusAccepted,
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

const baseUserProfileURL = `/api/users/profile/`

func TestUserProfileHandler(t *testing.T) {
	ctx := t.Context()

	app := tu.RouterSetup(ctx)
	tokenString := tu.LoginSetup(app, "vendor_user", "securePassword")

	// testing userProfileHandler.upsert
	w := httptest.NewRecorder()
	body := bytes.NewBuffer([]byte(`{"birthday": "2005-08-26", "bio": "some info..."}`))
	req, _ := http.NewRequest(http.MethodPost, baseUserProfileURL, body)
	req.Header.Set("Cookie", fmt.Sprintf("__Host-auth-token=%s", strings.TrimPrefix(tokenString, "Bearer ")))
	app.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf(`expected status="%d", but got "%d"`, http.StatusAccepted, w.Code)
	}

	// testing userProfileHandler.uploadProfileImg
	_, p, _, _ := runtime.Caller(0)
	basePath := filepath.Join(filepath.Dir(p), "..", "..", "internal", "testutils", "testfile", "temp.jpeg")
	w = httptest.NewRecorder()
	req, err := tu.FileUploadRequest(fmt.Sprintf("%supload", baseUserProfileURL), tokenString, "file", basePath)
	if err != nil {
		t.Errorf(`failed to create multipart request, %s`, err)
		return
	}
	app.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf(`expected status="%d", but got "%d"`, http.StatusAccepted, w.Code)
	}

	log := logger.GetLogger()
	us := queries.NewUserStore(database.GetSession(), log)
	user, err := us.Get(ctx, "vendor_user")
	if err != nil {
		t.Errorf("failed to get user, %s", err)
	}

	ups := queries.NewUserProfileStore(database.GetSession(), log)
	imgPath, err := ups.GetImgPath(ctx, user.ID)
	if err != nil || imgPath == "" {
		t.Errorf(`failed to save img in database, imgPath="%s", %s`, imgPath, err)
	}
	config := internal.GetConfig()
	if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Opt.UploadPath, imgPath)); err != nil {
		t.Errorf(`failed to save file "%s", %s`, imgPath, err)
	}

	// testing userProfileHandler.deleteImgFile
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, baseUserProfileURL, nil)
	req.Header.Set("Authorization", tokenString)
	app.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf(`expected status="%d", but got "%d"`, http.StatusNoContent, w.Code)
	}
	if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Opt.UploadPath, imgPath)); err == nil {
		t.Errorf("failed to remove img file")
	}
}
