package app

import (
	"context"
	"fmt"
	"generic-shop-sample/db"
	_ "generic-shop-sample/docs"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/background"
	md "generic-shop-sample/middlewares"
	"generic-shop-sample/routes"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	ctx    context.Context
	Router *gin.Engine
	config *internal.Config
}

// @title						Generic Shop API
// @version					0.1.0
// @host						localhost:8080
// @BasePath					/api
// @securityDefinitions.apikey	CookieAuth
// @in							cookie
// @name						__Host-auth-token
// @authorizationurl			http://localhost/api/auth/login
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @authorizationurl			http://localhost/api/auth/login
func NewApp(ctx context.Context, config *internal.Config) *App {
	gin.DisableConsoleColor()
	gin.SetMode(config.Mode)
	router := gin.New()
	router.MaxMultipartMemory = config.MaxMultipartMemory << 20
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		panic(fmt.Errorf("failed to set trusted proxies, %s", err))
	}

	logLevel := slog.LevelInfo
	if gin.Mode() != gin.ReleaseMode {
		logLevel = slog.LevelDebug
	}
	internal.InitAppLogger(logLevel, config.AppLoggerFilepath)

	setMiddlewares(ctx, router, config)
	internal.SetCustomValidators()
	setRoutes(ctx, router)

	return &App{
		ctx:    ctx,
		Router: router,
		config: config,
	}
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
			slog.Error("failed to listen and server", "error", err)
		}
	}()

	background.StartTasks(a.ctx, db.NewSession())

	<-a.ctx.Done()

	shutdown, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdown); err != nil {
		slog.Warn("failed to shutdown server", "error", err)
	}
}

func setMiddlewares(ctx context.Context, router *gin.Engine, config *internal.Config) {
	corsConfig := &md.CorsConfig{
		Origins:     config.Origins,
		Credentials: true,
		Methods:     []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
	}

	rl := md.NewRateLimiter(ctx, 500, 10*time.Minute, 30*time.Minute)

	router.Use(
		md.RequestLoggerMiddleware(config.RequestLoggerFilepath),
		gin.Recovery(),
		md.SecurityHeadersMiddleware(),
		rl.RateLimiterMiddleware(),
		md.CorsMiddleware(corsConfig),
	)
}

func setRoutes(ctx context.Context, router *gin.Engine) {
	routes.APIRouter(ctx, router.Group("/api"))
	if gin.Mode() == gin.DebugMode {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
}
