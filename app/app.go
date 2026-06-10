package app

import (
	"context"
	"fmt"
	"generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/app/routes"
	"generic-shop-sample/internal"
	"generic-shop-sample/storage/database"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	ctx    context.Context
	Router *gin.Engine
	config *internal.Config
}

func NewApp(ctx context.Context, config *internal.Config) *App {
	gin.DisableConsoleColor()
	gin.SetMode(config.Opt.Mode)
	router := gin.New()
	router.MaxMultipartMemory = config.Opt.MaxMultipartMemory << 20
	if err := router.SetTrustedProxies(config.Opt.TrustedProxies); err != nil {
		panic(fmt.Errorf("failed to set trusted proxies, %s", err))
	}

	logLevel := slog.LevelInfo
	if gin.Mode() != gin.ReleaseMode {
		logLevel = slog.LevelDebug
	}
	internal.InitAppLogger(logLevel, config.Opt.AppLoggerFilepath)

	setMiddlewares(ctx, router, config)
	internal.SetCustomValidators()
	setRoutes(ctx, config, router)

	return &App{
		ctx:    ctx,
		Router: router,
		config: config,
	}
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
			slog.Error("failed to listen and server", "error", err)
		}
	}()

	background.StartTasks(a.ctx, database.GetSession())

	<-a.ctx.Done()

	shutdown, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdown); err != nil {
		slog.Warn("failed to shutdown server", "error", err)
	}
}

func setMiddlewares(ctx context.Context, router *gin.Engine, config *internal.Config) {
	corsConfig := &md.CorsConfig{
		Origins:     config.Opt.Origins,
		Credentials: true,
		Methods:     []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
	}

	rl := md.NewRateLimiter(ctx, 500, 10*time.Minute, 30*time.Minute)

	router.Use(
		md.RequestLoggerMiddleware(config.Opt.RequestLoggerFilepath),
		gin.Recovery(),
		md.SecurityHeadersMiddleware(),
		rl.RateLimiterMiddleware(),
		md.CorsMiddleware(corsConfig),
	)
}

func setRoutes(ctx context.Context, config *internal.Config, router *gin.Engine) {
	routes.APIRouter(ctx, router.Group("/api"))
	routes.StaticRouter(config, router.Group("/static"))
}
