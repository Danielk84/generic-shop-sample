package app

import (
	"context"
	md "generic-shop-sample/middlewares"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	ctx    context.Context
	Router *gin.Engine
	config *Config
}

func NewApp(ctx context.Context, config *Config) *App {
	gin.DisableConsoleColor()
	gin.SetMode(config.Mode)
	router := gin.Default()
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		log.Panicln(err)
	}

	setMiddlewares(ctx, router, config)

	return &App{
		ctx:    ctx,
		Router: router,
		config: config,
	}
}

func setMiddlewares(ctx context.Context, router *gin.Engine, config *Config) {
	corsConfig := &md.CorsConfig{
		Origins:     config.Origins,
		Credentials: true,
		Methods:     []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
	}

	rl := md.NewRateLimiter(ctx, 500, 10*time.Minute, 30*time.Minute)

	router.Use(
		md.SecurityHeadersMiddleware(),
		rl.RateLimiterMiddleware(),
		md.CorsMiddleware(corsConfig),
	)
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:         a.config.Addr,
		Handler:      a.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-a.ctx.Done()

	shutdown, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdown); err != nil {
		log.Println("error shutting down server: ", err)
	}
}
