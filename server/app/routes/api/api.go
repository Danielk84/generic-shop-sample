package api

import (
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

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

func GenFileKey(claims *auth.AuthClaims, dst string) string {
	y, m, d := time.Now().Date()
	return fmt.Sprintf("%d/%d/%d/%s-%d", y, m, d, claims.ID, time.Now().UnixNano())
}

func JSONResponse(c *gin.Context, code int, msg, defaultMsg string) {
	if msg == "" {
		msg = defaultMsg
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
	minN := 100000
	maxN := 999999
	return rand.Intn(maxN-minN) + minN
}

func SetJSONCache(ctx context.Context, cache cache.CacheClient, key string, expiration time.Duration, value any) error {
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

type JsonCacheInput[T any] struct {
	Ctx        context.Context
	CacheKey   string
	Client     cache.CacheClient
	Expiration time.Duration
	Log        logger.Logger
	Fn         func(context.Context) (T, error)
}

func JsonCache[T any](j JsonCacheInput[T]) (output T, err error) {
	if err = GetJSONCache(j.Ctx, j.Client, j.CacheKey, &output); err != nil {
		LogCacheErr("GetJSONCache", j.CacheKey, err)

		if output, err = j.Fn(j.Ctx); err != nil {
			return
		}
		if err = SetJSONCache(j.Ctx, j.Client, j.CacheKey, j.Expiration, output); err != nil {
			LogCacheErr("SetJSONCache", j.CacheKey, err)
			err = nil
		}
	}
	return
}

type CacheMaxPageInput struct {
	ctx        context.Context
	client     cache.CacheClient
	name       string
	pagination int
	getMaxPage queries.MaxPageType
}

func CacheMaxPage(c CacheMaxPageInput) (count int, err error) {
	cacheKey := fmt.Sprintf("max-page:%s", c.name)
	if err = c.client.Get(c.ctx, cacheKey).Scan(&count); err != nil {
		LogCacheErr("Get", "CacheMaxPage", err)

		if count, err = c.getMaxPage(c.ctx, c.pagination); err != nil {
			return
		}
		if err = c.client.Set(c.ctx, cacheKey, count, time.Hour).Err(); err != nil {
			LogCacheErr("Set", "CacheMaxPage", err)
			err = nil
		}
	}
	return
}

func SetPageHeader(c *gin.Context, maxPage CacheMaxPageInput) {
	count, err := CacheMaxPage(maxPage)
	if err != nil {
		return
	}
	c.Header("X-Max-Page", strconv.Itoa(count))
}
