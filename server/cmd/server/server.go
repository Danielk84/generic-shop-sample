package main

import (
	"fmt"
	"generic-shop-sample/app"
	bg "generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/app/routes"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/cache_query"
	"generic-shop-sample/storage/queries"
	"time"

	"github.com/gin-gonic/gin"
)

type server struct {
	app  *app.App
	deps *app.ServiceDeps
}

func newServer(deps *app.ServiceDeps) *server {
	sv := &server{
		app:  app.NewApp(deps.Ctx, deps.Config.App),
		deps: deps,
	}
	log := logger.GetLogger()
	log.Debug("newServer", "config", fmt.Sprintf("%+v", deps.Config))
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
	rl := md.NewRateLimiter(s.deps.Ctx, 500, 10*time.Minute, 30*time.Minute)

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
			Log:      log,
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
		&bg.SearchIndexProcess{
			Store: queries.NewSearchStore(session, log),
			Log:   log,
		},
	}
	for _, t := range tasks {
		go t.Start(s.deps.Ctx)
	}
}
