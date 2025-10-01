package app

import (
	"context"
	"generic-shop-sample/internal"
	md "generic-shop-sample/middlewares"
	"generic-shop-sample/routes"
	"log"
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
	gin.SetMode(config.Mode)
	router := gin.Default()
	router.MaxMultipartMemory = config.MaxMultipartMemory << 20
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		log.Panicln(err)
	}

	setMiddlewares(ctx, router, config)
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

func setMiddlewares(ctx context.Context, router *gin.Engine, config *internal.Config) {
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

func setRoutes(ctx context.Context, router *gin.Engine) {
	routes.APIRouter(ctx, router.Group("/api"))
}
