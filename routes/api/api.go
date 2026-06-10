package api

import (
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/db/cache"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"io"
	"log/slog"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var defaultPagination = internal.GetConfig().Opt.Pagination

type RouteSpec struct {
	Method       string
	RelativePath string
	Handlers     []gin.HandlerFunc
}

func RegisterRoutesWith(router *gin.RouterGroup, middlewares []gin.HandlerFunc, ehs []RouteSpec) {
	for _, eh := range ehs {
		handlers := append(middlewares, eh.Handlers...)
		router.Handle(eh.Method, eh.RelativePath, handlers...)
	}
}

type SetFlag struct {
	Accepted bool `json:"accepted" binding:"boolean"`
}

func GetPage(c *gin.Context) int {
	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		return 1
	}
	return page
}

func UploadFile(file *multipart.FileHeader, claims *auth.AuthClaims, group, dst string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file")
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	if _, err = src.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	mtype := mimetype.Detect(buf[:n])
	config := internal.GetConfig()
	if !mimetype.EqualsAny(mtype.String(), config.Opt.AllowedImgMimetype...) {
		return "", err
	}

	if dst == "" {
		y, m, d := time.Now().Date()
		dst = fmt.Sprintf("%s/%d/%d/%d/%s-%d%s", group, y, m, d, claims.Username, time.Now().UnixNano(), mtype.Extension())
	}
	path := fmt.Sprintf("%s/%s", config.Opt.UploadPath, dst)
	if err = os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return "", err
	}
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	return dst, nil
}

func JSONResponse(c *gin.Context, code int, msg, DefaultMsg string) {
	if msg == "" {
		msg = DefaultMsg
	}
	c.JSON(code, gin.H{"msg": msg})
}

func Created(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusCreated, msg, "Items created successfully.")
}

func Accepted(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusAccepted, msg, "Items accepted successfully.")
}

func NotFound(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusNotFound, msg, "Page Not Found!")
}

func BadRequest(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusBadRequest, msg, "Invalid request")
}

func Unauthorized(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusUnauthorized, msg, "Authorization required!")
}

func Forbidden(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusForbidden, msg, "You don't have permission to access this resource.")
}

func Unprocessable(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusUnprocessableEntity, msg, "Failed to process entity")
}

func HasPermissions(c *gin.Context, userPermission queries.PermissionType, permissions ...queries.PermissionType) bool {
	isContains := slices.Contains(permissions, userPermission)
	if !isContains && c != nil {
		Forbidden(c, "")
	}
	return isContains
}

func RandVerifyNum() int {
	min := 100000
	max := 999999
	return rand.Intn(max-min) + min
}

func SetJSONCacheEx(ctx context.Context, cache cache.CacheClient, key string, expiration time.Duration, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = cache.SetEx(ctx, key, data, expiration).Result()
	return err
}

func GetJSONCache[T any](ctx context.Context, cache cache.CacheClient, key string, value *T) error {
	data, err := cache.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), value)
	return err
}

func LogCacheErr(method, section string, err error) {
	if err == redis.Nil {
		return
	}
	slog.Error("failed to process cache",
		"method", method,
		"secion", section,
		"err", err,
	)
}
