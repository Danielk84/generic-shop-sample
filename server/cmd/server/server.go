package main

import (
	"context"
	"generic-shop-sample/app"
	bg "generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/app/routes"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/cache_query"
	"generic-shop-sample/storage/queries"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type server struct {
	ctx  context.Context
	app  *app.App
	deps *app.ServiceDeps
}

func newServer(ctx context.Context, deps *app.ServiceDeps) *server {
	sv := &server{
		ctx:  ctx,
		app:  app.NewApp(ctx, deps.Config.App),
		deps: deps,
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
	rlogWriter := logger.CreateLogFile(s.deps.Config.RequestLoggerFilepath)
	s.app.OpenWriter = append(s.app.OpenWriter, rlogWriter)
	rl := md.NewRateLimiter(s.ctx, 500, 10*time.Minute, 30*time.Minute)

	s.app.Router.Use(
		md.RequestLoggerMiddleware(rlogWriter),
		gin.Recovery(),
		rl.RateLimiterMiddleware(),
	)
}

func (s *server) setRoutes() {
	routes.APIRouter(s.deps, s.app.Router.Group("/api"))
	routes.StaticRouter(s.deps, s.app.Router.Group("/static"))
}

func (s *server) setBackgroundTask() {
	log := logger.GetLogger()
	session := s.deps.DB.GetSession()
	tasks := []bg.BackgroundTask{
		&bg.ExpiredOrdersCleaner{
			Store: queries.NewOrderStore(session, log),
			Log:   log,
		},
		&bg.EmailBroker{
			Cache:    s.deps.Cache.GetCache(cache.UsersCache),
			From:     s.deps.Config.EmailBroker.FromEmail,
			Host:     s.deps.Config.EmailBroker.SMTPHost,
			Port:     s.deps.Config.EmailBroker.SMTPPort,
			Password: s.deps.Config.EmailBroker.SMTPPassword,
		},
		&bg.OrdersProcess{
			Cache:          s.deps.Cache.GetCache(cache.OrdersCache),
			Log:            log,
			OrderStore:     queries.NewOrderStore(session, log),
			OrderItemStore: queries.NewOrderItemsStore(session, log),
			ProductStore:   queries.NewProductStore(session, log),
			VendorStore:    queries.NewVendorOrderStore(session, log),
			VendorCacheStore: cache_query.NewVendorCacheStore(
				s.deps.Cache.GetCache(cache.OrdersCache),
				log,
				queries.NewProductStore(session, log)),
			Pagination: s.deps.Config.Pagination,
		},
	}
	for _, t := range tasks {
		go t.Start(s.ctx)
	}
}
