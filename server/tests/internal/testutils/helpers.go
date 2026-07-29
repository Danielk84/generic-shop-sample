package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/app"
	"generic-shop-sample/internal/config"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type TestFixtureFn func(deps *app.ServiceDeps) error

func ConfigTestSetup() config.Config {
	configFile, ok := os.LookupEnv("TEST_CONFIG")
	if !ok {
		panic(`Please set "TEST_CONFIG=/path/to/config.yaml"`)
	}
	return config.NewConfig(configFile)
}

func DBTestSetup(ctx context.Context, config config.Config) database.DBManager {
	db, err := database.New(ctx, config.DatabaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to setup db, %s", err))
	}
	return db
}

func CacheTestSetup(ctx context.Context, config config.Config) cache.CacheManager {
	c, err := cache.New(ctx, config.CacheURL,
		[]int{cache.PublicCache,
			cache.UsersCache,
			cache.ProductsCache,
			cache.OrdersCache,
			cache.PaymentCache})
	if err != nil {
		panic(fmt.Errorf("failed to setup cache, %s", err))
	}
	return c
}

func LogTestSetup(config config.Config) logger.Logger {
	logWriter := logger.CreateLogFile(config.App.AppLoggerFilepath)
	return logger.SetLogger(logger.LevelDebug, logWriter)
}

func ServiceDepsTestSetup(ctx context.Context, config config.Config) *app.ServiceDeps {
	return &app.ServiceDeps{
		Ctx:    ctx,
		Config: config,
		DB:     DBTestSetup(ctx, config),
		Cache:  CacheTestSetup(ctx, config),
	}
}

func CloseServieDepsTestSetup(deps *app.ServiceDeps) {
	deps.DB.Close()
	deps.Cache.Close()
}

func CeckErrList(name string, errs []error) {
	log := logger.GetLogger()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Error("testMain - "+name, "error", err)
		}
		os.Exit(1)
	}
}

func LoginSetup(app *gin.Engine, username, password string) string {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
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

func CreateTemporaryUsers(deps *app.ServiceDeps) error {
	batch := &pgx.Batch{}
	const q = `INSERT INTO user_s.users (username, password, permission_type, is_active)
		VALUES (@Username, @Password, @PermissionType, @IsActive)`
	password := "secure password"
	var users = []queries.CreateUserRequest{
		{
			LoginRequest:          queries.LoginRequest{Username: "adminUser", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Admin, IsActive: true}},
		{
			LoginRequest:          queries.LoginRequest{Username: "vendorUser", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Vendor, IsActive: true}},
		{
			LoginRequest:          queries.LoginRequest{Username: "customerUser", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Customer, IsActive: true}},
		{
			LoginRequest:          queries.LoginRequest{Username: "blockUser", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.BlockUser, IsActive: false}},
	}

	for _, user := range users {
		args := pgx.NamedArgs{
			"Username":       user.Username,
			"Password":       user.Password,
			"PermissionType": user.PermissionType,
			"IsActive":       user.IsActive,
		}
		batch.Queue(q, args)
	}

	sb := deps.DB.GetSession().SendBatch(deps.Ctx, batch)
	return sb.Close()
}

func TruncateTables(deps *app.ServiceDeps) error {
	const q = `TRUNCATE
			order_s.order_items,
			order_s.orders,
			product_s.products,
			product_s.categories,
			order_s.vendors_order,
			user_s.users
		RESTART IDENTITY CASCADE`
	_, err := deps.DB.GetSession().Exec(deps.Ctx, q)
	return err
}

func FlushCache(deps *app.ServiceDeps) error {
	namespace := []int{
		cache.PublicCache,
		cache.UsersCache,
		cache.ProductsCache,
		cache.OrdersCache,
		cache.PaymentCache,
	}
	errs := []error{}
	for _, n := range namespace {
		if err := deps.Cache.GetCache(n).FlushAll(deps.Ctx).Err(); err != nil {
			errs = append(errs, err)
		}
	}
	CeckErrList("FlushAllCache", errs)
	return nil
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
