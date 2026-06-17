package app

import (
	"context"
	"fmt"
	"generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/app/routes"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/logger"
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
	config     *internal.Config
	log        logger.Logger
	openWriter []io.WriteCloser
}

func NewApp(ctx context.Context, config *internal.Config) *App {
	gin.DisableConsoleColor()
	gin.SetMode(config.Opt.Mode)

	app := &App{ctx: ctx, Router: gin.New()}
	app.Router.MaxMultipartMemory = config.Opt.MaxMultipartMemory << 20
	if err := app.Router.SetTrustedProxies(config.Opt.TrustedProxies); err != nil {
		panic(fmt.Errorf("failed to set trusted proxies, %s", err))
	}

	logLevel := logger.LevelWarn
	if gin.Mode() != gin.ReleaseMode {
		logLevel = logger.LevelDebug
	}
	logWriter := logger.CreateLogFile(config.Opt.AppLoggerFilepath)
	app.openWriter = append(app.openWriter, logWriter)
	app.config = config
	app.log = logger.SetLogger(logLevel, logWriter)

	app.setMiddlewares()
	internal.SetCustomValidators()
	app.setRoutes()

	return app
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:              a.config.Opt.Addr,
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

	background.StartTasks(a.ctx, database.GetSession())

	<-a.ctx.Done()

	shutdown, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdown); err != nil {
		a.log.Warn("failed to shutdown server", "error", err)
	}

	a.Close()
}

func (a *App) Close() {
	for _, writer := range a.openWriter {
		if err := writer.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close log writer: %s", err)
		}
	}
}

func (a *App) setMiddlewares() {
	rlogWriter := logger.CreateLogFile(a.config.Opt.RequestLoggerFilepath)
	a.openWriter = append(a.openWriter, rlogWriter)

	corsConfig := &md.CorsConfig{
		Origins:     a.config.Opt.Origins,
		Credentials: true,
		Methods:     []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
	}

	rl := md.NewRateLimiter(a.ctx, 500, 10*time.Minute, 30*time.Minute)

	a.Router.Use(
		md.RequestLoggerMiddleware(rlogWriter),
		gin.Recovery(),
		md.SecurityHeadersMiddleware(),
		rl.RateLimiterMiddleware(),
		md.CorsMiddleware(corsConfig),
	)
}

func (a *App) setRoutes() {
	routes.APIRouter(a.ctx, a.Router.Group("/api"))
	routes.StaticRouter(a.config, a.Router.Group("/static"))
}
