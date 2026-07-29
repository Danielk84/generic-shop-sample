package api

import (
	"generic-shop-sample/app"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/file_storage"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func FileRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := fileHandler{
		fileStore: file_storage.NewFileStore(deps.Ctx, deps.Config.FileStore, deps.FileStore, "global-static"),
		log:       log,
		mimetypes: deps.Config.FileStore.AllowedImgMimetype,
	}
	router.GET("imgs/:bucket/*filepath", h.getImgs)
}

type fileHandler struct {
	fileStore file_storage.FileStore
	log       logger.Logger
	mimetypes []string
}

func (f *fileHandler) getImgs(c *gin.Context) {
	bucket := c.Param("bucket")
	filepath := c.Param("filepath")
	ctx := c.Request.Context()

	obj, err := f.fileStore.Download(ctx, bucket, filepath)
	if err != nil {
		NotFound(c, "content not found")
		return
	}
	defer func() { _ = obj.Body.Close() }()
	if obj.ContentLength == nil {
		NotFound(c, "content not found")
		f.log.Warn("fileHandler.getImgs",
			"error", "ContentLength is nil",
			"bucket", bucket,
			"filepath", filepath)
		return
	}

	mimetype := ""
	for _, m := range f.mimetypes {
		s := strings.Split(m, "/")
		if len(s) != 2 {
			continue
		}
		if _, found := strings.CutPrefix(filepath, s[1]); found {
			mimetype = m
			break
		}
	}
	if mimetype == "" {
		NotFound(c, "")
		f.log.Error("fileHandler.getImgs",
			"error", "failed on matching mimetype",
			"mimeTypes", f.mimetypes,
			"key", filepath)
		return
	}

	c.DataFromReader(http.StatusOK,
		*obj.ContentLength,
		mimetype,
		obj.Body,
		map[string]string{},
	)
}
