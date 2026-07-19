package app

import (
	"context"
	"fmt"
	"generic-shop-sample/internal/config"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	ctx        context.Context
	Router     *gin.Engine
	config     config.AppConfig
	log        logger.Logger
	OpenWriter []io.WriteCloser
}

func NewApp(ctx context.Context, config config.AppConfig) *App {
	gin.DisableConsoleColor()
	gin.SetMode(config.Mode)

	app := &App{ctx: ctx, Router: gin.New()}
	app.Router.MaxMultipartMemory = config.MaxMultipartMemory << 20
	if err := app.Router.SetTrustedProxies(config.TrustedProxies); err != nil {
		panic(fmt.Errorf("failed to set trusted proxies, %s", err))
	}

	logLevel := logger.LevelWarn
	if gin.Mode() != gin.ReleaseMode {
		logLevel = logger.LevelDebug
	}
	logWriter := logger.CreateLogFile(config.AppLoggerFilepath)
	app.OpenWriter = append(app.OpenWriter, logWriter)
	app.config = config
	app.log = logger.SetLogger(logLevel, logWriter)

	return app
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:              a.config.Addr,
		Handler:           http.TimeoutHandler(a.Router, 10*time.Second, "request timeout"),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       20 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("failed to listen and server", "error", err)
		}
	}()

	<-a.ctx.Done()

	shutdown, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdown); err != nil {
		a.log.Warn("failed to shutdown server", "error", err)
	}
}

func (a *App) Close() {
	for _, writer := range a.OpenWriter {
		if err := writer.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close log writer: %s", err)
		}
	}
}

type ServiceDeps struct {
	Ctx    context.Context
	DB     database.DBManager
	Cache  cache.CacheManager
	Config config.Config
}
