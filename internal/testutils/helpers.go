package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/app"
	"generic-shop-sample/db"
	"generic-shop-sample/db/database"
	"generic-shop-sample/internal"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func RouterSetup(ctx context.Context) *gin.Engine {
	config := internal.NewConfig()
	app := app.NewApp(ctx, config)

	return app.Router
}

func DBManagerSetup(ctx context.Context) database.DBManager {
	config := internal.NewConfig()
	db, err := database.New(ctx, config.DatabaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to setup db, %s", err))
	}
	return db
}

func CacheSetup(ctx context.Context) db.CacheManager {
	config := internal.NewConfig()
	cache, err := db.NewCacheManager(ctx, config.CacheURL, []int{db.PublicCache, db.UsersCache, db.ProductsCache})
	if err != nil {
		panic(fmt.Errorf("failed to setup cache, %s", err))
	}
	return cache
}

func CeckErrList(name string, errs []error) {
	if len(errs) > 0 {
		for _, err := range errs {
			slog.Error("testMain - "+name, "error", err)
		}
		os.Exit(1)
	}
}

func LoginSetup(app *gin.Engine, username, password string) string {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	app.ServeHTTP(w, req)
	var resJson map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resJson)
	return resJson["token"]
}

func FileUploadRequest(url, token, fileName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreateFormFile(fileName, filepath.Base(path))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", token)
	return req, err
}
