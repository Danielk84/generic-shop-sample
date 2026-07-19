package main

import (
	"context"
	"generic-shop-sample/app"
	bg "generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/app/routes"
	"generic-shop-sample/internal/config"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/cache_query"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type server struct {
	ctx    context.Context
	app    *app.App
	config config.Config
	db     database.DBManager
	cache  cache.CacheManager
}

func newServer(ctx context.Context, config config.Config, db database.DBManager, cache cache.CacheManager) *server {
	sv := &server{
		ctx:    ctx,
		app:    app.NewApp(ctx, config.App),
		config: config,
		db:     db,
		cache:  cache,
	}
	sv.setMiddlewares()
	sv.setRoutes()
	return sv
}

func (s *server) run() {
	s.setBackgroundTask()
	s.app.Run()
	s.app.Close()
}

func (s *server) setMiddlewares() {
	rlogWriter := logger.CreateLogFile(s.config.RequestLoggerFilepath)
	s.app.OpenWriter = append(s.app.OpenWriter, rlogWriter)

	corsConfig := &md.CorsConfig{
		Origins:     s.config.Origins,
		Credentials: true,
		Methods:     []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
	}

	rl := md.NewRateLimiter(s.ctx, 500, 10*time.Minute, 30*time.Minute)

	s.app.Router.Use(
		md.RequestLoggerMiddleware(rlogWriter),
		gin.Recovery(),
		md.SecurityHeadersMiddleware(),
		rl.RateLimiterMiddleware(),
		md.CorsMiddleware(corsConfig),
	)
}

func (s *server) setRoutes() {
	deps := &app.ServiceDeps{
		Ctx:    s.ctx,
		DB:     s.db,
		Cache:  s.cache,
		Config: s.config,
	}
	routes.APIRouter(deps, s.app.Router.Group("/api"))
	routes.StaticRouter(s.config, s.app.Router.Group("/static"))
}

func (s *server) setBackgroundTask() {
	log := logger.GetLogger()
	session := s.db.GetSession()
	tasks := []bg.BackgroundTask{
		&bg.ExpiredOrdersCleaner{
			Store: queries.NewOrderStore(session, log),
			Log:   log,
		},
		&bg.EmailBroker{
			Cache:    s.cache.GetCache(cache.UsersCache),
			From:     s.config.EmailBroker.FromEmail,
			Host:     s.config.EmailBroker.SMTPHost,
			Port:     s.config.EmailBroker.SMTPPort,
			Password: s.config.EmailBroker.SMTPPassword,
		},
		&bg.OrdersProcess{
			Cache:          s.cache.GetCache(cache.OrdersCache),
			Log:            log,
			OrderStore:     queries.NewOrderStore(session, log),
			OrderItemStore: queries.NewOrderItemsStore(session, log),
			ProductStore:   queries.NewProductStore(session, log),
			VendorStore:    queries.NewVendorOrderStore(session, log),
			VendorCacheStore: cache_query.NewVendorCacheStore(
				s.cache.GetCache(cache.OrdersCache),
				log,
				queries.NewProductStore(session, log)),
			Pagination: s.config.Pagination,
		},
	}
	for _, t := range tasks {
		go t.Start(s.ctx)
	}
}
