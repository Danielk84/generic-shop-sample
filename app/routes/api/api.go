package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"io"
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

var (
	ErrReadFile        = errors.New("failed to read file")
	ErrOperationFailed = errors.New("operation failed")
	ErrInvalidMimeType = errors.New("invalid mime type")
	ErrUploadFile      = errors.New("failed to upload and save file")
)

var defaultPagination = internal.GetConfig().Opt.Pagination

type RouteSpec struct {
	Method       string
	RelativePath string
	Handlers     []gin.HandlerFunc
}

func RegisterRoutesWith(router *gin.RouterGroup, middlewares []gin.HandlerFunc, rs []RouteSpec) {
	if len(middlewares) == 0 || len(rs) == 0 {
		panic("invalid empty middlewares or RouteSpec list")
	}
	for _, e := range rs {
		if e.Method == "" || len(e.Handlers) == 0 {
			panic("invalid RouteSpec")
		}
		handlers := append(middlewares, e.Handlers...)
		router.Handle(e.Method, e.RelativePath, handlers...)
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

type FileUploaderFunc func(file *multipart.FileHeader, claims *auth.AuthClaims, group, dst string) (string, error)

func GetFileUploader(config *internal.Config, log logger.Logger) FileUploaderFunc {
	uf := uploadFile{config, log}
	return uf.local
}

type uploadFile struct {
	config *internal.Config
	log    logger.Logger
}

func (u *uploadFile) local(file *multipart.FileHeader, claims *auth.AuthClaims, group, dst string) (string, error) {
	src, err := file.Open()
	if err != nil {
		u.log.Warn("uploadFile.local", "step", "file.open", "error", err)
		return "", ErrReadFile
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil {
		u.log.Error("uploadFile.local", "step", "src.Read", "error", err)
		return "", ErrReadFile
	}
	if _, err = src.Seek(0, io.SeekStart); err != nil {
		u.log.Error("uploadFile.local", "step", "src.Seek", "error", err)
		return "", ErrOperationFailed
	}
	mtype := mimetype.Detect(buf[:n])
	if !mimetype.EqualsAny(mtype.String(), u.config.Opt.AllowedImgMimetype...) {
		return "", ErrInvalidMimeType
	}

	if dst == "" {
		y, m, d := time.Now().Date()
		dst = fmt.Sprintf("%s/%d/%d/%d/%s-%d%s", group, y, m, d, claims.Username, time.Now().UnixNano(), mtype.Extension())
	}
	path := fmt.Sprintf("%s/%s", u.config.Opt.UploadPath, dst)
	u.log.Debug("uploadFile.local", "path", path)
	if err = os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		u.log.Error("uploadFile.local", "step", "os.MkdirAll", "error", err)
		return "", ErrOperationFailed
	}
	out, err := os.Create(path)
	if err != nil {
		u.log.Error("uploadFile.local", "step", "os.Create", "error", err)
		return "", ErrOperationFailed
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		u.log.Error("uploadFile.local", "step", "io.Copy", "error", err)
		return "", ErrUploadFile
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
	_, err = cache.Set(ctx, key, data, expiration).Result()
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
	logger.GetLogger().Error("failed to process cache",
		"method", method,
		"secion", section,
		"err", err,
	)
}
