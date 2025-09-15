package app

import (
	"context"
	md "generic-shop-sample/middlewares"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
	config *AppConfig
}

func NewApp(config *AppConfig) *App {
	gin.DisableConsoleColor()
	gin.SetMode(config.Mode)
	router := gin.Default()
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		log.Panicln(err)
	}

	setMiddlewares(router)

	return &App{
		Router: router,
		config: config,
	}
}

func setMiddlewares(router *gin.Engine) {
	corsConfig := &md.CorsConfig{
		Origins:     []string{"http://localhost:8080/"},
		Credentials: true,
		Methods:     []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
	}
	rl := md.NewRateLimiter(500, 10)

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

	quit, quitCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer quitCancel()
	<-quit.Done()

	shutdown, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdown); err != nil {
		log.Println("error shutting down server: ", err)
	}
}
