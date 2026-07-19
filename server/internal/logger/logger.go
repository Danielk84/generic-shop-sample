package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type Level = slog.Level

var (
	LevelDebug = slog.LevelDebug
	LevelWarn  = slog.LevelWarn
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type logWriteCloser struct {
	writer io.Writer
	closer io.Closer
}

func (l *logWriteCloser) Write(p []byte) (int, error) {
	return l.writer.Write(p)
}

func (l *logWriteCloser) Close() error {
	return l.closer.Close()
}

func CreateLogFile(fp string) io.WriteCloser {
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, 0750); err != nil {
		panic(fmt.Errorf(`failed to mkdir "%s", %s\n`, dir, err))
	}
	file, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		panic(fmt.Errorf(`failed to create file "%s", %s`, fp, err))
	}
	lwc := &logWriteCloser{closer: file, writer: file}
	if gin.Mode() != gin.ReleaseMode {
		lwc.writer = io.MultiWriter(os.Stdout, file)
	}
	return lwc
}

func SetLogger(level Level, writer io.Writer) Logger {
	programLevel := new(slog.LevelVar)
	programLevel.Set(level)
	logger := slog.New(slog.NewJSONHandler(
		writer,
		&slog.HandlerOptions{Level: programLevel},
	))
	slog.SetDefault(logger)
	return logger
}

func GetLogger() Logger {
	return slog.Default()
}
