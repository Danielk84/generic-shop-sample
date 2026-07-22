package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/app"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func GetConfig() *internal.Config {
	configFile, ok := os.LookupEnv("TEST_CONFIG")
	if !ok {
		panic(`Please set "TEST_CONFIG=/path/to/config.yaml"`)
	}
	return internal.NewConfig(configFile)
}

func RouterSetup(ctx context.Context) *gin.Engine {
	config := GetConfig()
	app := app.NewApp(ctx, config)

	return app.Router
}

func DBManagerSetup(ctx context.Context) database.DBManager {
	config := GetConfig()
	db, err := database.New(ctx, config.Opt.DatabaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to setup db, %s", err))
	}
	return db
}

func CacheSetup(ctx context.Context) cache.CacheManager {
	config := GetConfig()
	cache, err := cache.New(ctx, config.Opt.CacheURL, []int{cache.PublicCache, cache.UsersCache, cache.ProductsCache})
	if err != nil {
		panic(fmt.Errorf("failed to setup cache, %s", err))
	}
	return cache
}

func CeckErrList(name string, errs []error) {
	if len(errs) > 0 {
		for _, err := range errs {
			logger.GetLogger().Error("testMain - "+name, "error", err)
		}
		os.Exit(1)
	}
}

func LoginSetup(app *gin.Engine, username, password string) string {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	app.ServeHTTP(w, req)
	cokies := w.Result().Cookies()
	return cokies[0].Value
}

func AddAuthCookie(req *http.Request, authToken string) {
	cookie := http.Cookie{}
	cookie.Name = "__Host-auth-token"
	cookie.Value = authToken
	req.AddCookie(&cookie)
}

func FileUploadRequest(url, token, fileName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer func() { _ = writer.Close() }()

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
	AddAuthCookie(req, token)
	return req, err
}
