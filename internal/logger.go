package internal

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func CreateLogFile(fp string) io.Writer {
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, 0750); err != nil {
		panic(fmt.Errorf(`failed to mkdir "%s", %s\n`, dir, err))
	}
	file, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		panic(fmt.Errorf(`failed to create file "%s", %s`, fp, err))
	}
	if gin.Mode() != gin.ReleaseMode {
		return io.MultiWriter(os.Stdout, file)
	}
	return file
}

func InitAppLogger(level slog.Level, fp string) {
	programLevel := new(slog.LevelVar)
	programLevel.Set(level)
	logger := slog.New(slog.NewJSONHandler(
		CreateLogFile(fp),
		&slog.HandlerOptions{Level: programLevel},
	))
	slog.SetDefault(logger)
}
