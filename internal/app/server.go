package app

import (
	"os"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"
	"inshorts.com/inshorts-news-srv/internal/base"
	"inshorts.com/inshorts-news-srv/internal/news"
	"inshorts.com/inshorts-news-srv/internal/router"
	"inshorts.com/inshorts-news-srv/internal/trending"
)

type App struct {
	server *fiber.App
}

func New() *App {
	return &App{
		server: fiber.New(),
	}
}

func (app *App) Start(log *zerolog.Logger, port string, es *elasticsearch.Client, newsService *news.Service, trendingSvc *trending.TrendingService, articles map[string]news.Article) {
	// set 1 MB headersize
	app.server.Server().ReadBufferSize = 1 * 1024 * 1024

	// CORS
	cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, ApiKey",
	})

	// Prometheus middleware to collect metrics
	prometheus := fiberprometheus.New(base.ServiceName)
	prometheus.RegisterAt(app.server, "/metrics")
	app.server.Use(prometheus.Middleware)
	log.Info().Caller().Msg("/metrics added...")

	// Context path
	contextPath := os.Getenv(base.EnvContextPath)
	if contextPath == "" {
		contextPath = base.ServiceName // Fallback context path to service name if ENV variable is not set
	}

	routes := app.server.Group(contextPath)
	router.CreateRoutes(routes, es, newsService, trendingSvc, articles)

	log.Info().Msgf("service listening on :%s", port)
	err := app.server.Listen(":" + port)
	if err != nil {
		log.Fatal().Err(err).Stack().Msgf("server failed: %v ", err.Error())
	}
}

func (app *App) Stop() error {
	return app.server.Shutdown()
}
