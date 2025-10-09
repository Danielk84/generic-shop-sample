package api

import (
	"fmt"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

type SetFlag struct {
	Accepted bool `json:"accepted"`
}

var defaultPagination = internal.NewConfig().Pagination

func getOffsetFromPageNum(p string) int {
	page, err := strconv.Atoi(p)
	if err != nil || page < 1 {
		return 1
	}
	return (page - 1) * defaultPagination
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
	config := internal.NewConfig()
	if !mimetype.EqualsAny(mtype.String(), config.AllowedImgMimetype...) {
		return "", err
	}

	if dst == "" {
		y, m, d := time.Now().Date()
		dst = fmt.Sprintf("%s/%s/%d/%d/%d/%s-%d%s", config.UploadPath, group, y, m, d, claims.Username, time.Now().UnixNano(), mtype.Extension())
	}
	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return "", err
	}
	out, err := os.Create(dst)
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
	c.JSON(code, msg)
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
	JSONResponse(c, http.StatusBadRequest, msg, "invalid request")
}

func Unauthorized(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusUnauthorized, msg, "Authorization required!")
}

func Forbidden(c *gin.Context, msg string) {
	JSONResponse(c, http.StatusForbidden, msg, "You don't have permission to access this resource.")
}

func HasPermissions(c *gin.Context, userPermission queries.PermissionType, permissions ...queries.PermissionType) bool {
	isContains := slices.Contains(permissions, userPermission)
	if !isContains && c != nil {
		Forbidden(c, "")
	}
	return isContains
}
