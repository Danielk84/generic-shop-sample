package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	router *gin.Engine
	config *AppConfig
}

func NewApp(config *AppConfig) *App {
	gin.SetMode(config.Mode)
	router := gin.Default()
	router.SetTrustedProxies(config.TrustedProxies)

	return &App{
		router: router,
		config: config,
	}
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:    a.config.Addr,
		Handler: a.router,
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
