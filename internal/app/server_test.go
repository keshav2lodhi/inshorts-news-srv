package app

import (
	"net/http"
	"os"
	"testing"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gofiber/fiber/v2"

	"github.com/stretchr/testify/assert"

	"inshorts.com/inshorts-news-srv/internal/base"
	"inshorts.com/inshorts-news-srv/internal/news"
	"inshorts.com/inshorts-news-srv/internal/router"
	"inshorts.com/inshorts-news-srv/internal/trending"
)

func setupTestApp(
	t *testing.T,
	contextPath string,
) *fiber.App {
	t.Helper()

	if contextPath != "" {
		os.Setenv(base.EnvContextPath, contextPath)
	} else {
		os.Unsetenv(base.EnvContextPath)
	}

	app := New()

	// Fake deps (safe because handlers won’t execute deep logic)
	var es *elasticsearch.Client
	var newsSvc *news.Service
	trendingSvc := trending.NewTrendingService(nil)

	articles := map[string]news.Article{}

	// --- inline copy of Start() without Listen ---
	app.server.Server().ReadBufferSize = 1 * 1024 * 1024

	app.server.Use(func(c *fiber.Ctx) error {
		return c.Next()
	})

	prom := fiber.New()
	_ = prom

	// Prometheus
	p := fiberprometheus.New(base.ServiceName)
	p.RegisterAt(app.server, "/metrics")
	app.server.Use(p.Middleware)

	ctxPath := contextPath
	if ctxPath == "" {
		ctxPath = base.ServiceName
	}

	routes := app.server.Group(ctxPath)
	router.CreateRoutes(routes, es, newsSvc, trendingSvc, articles)

	return app.server
}

func TestNewApp(t *testing.T) {
	app := New()
	assert.NotNil(t, app)
	assert.NotNil(t, app.server)
}

func TestMetricsEndpoint(t *testing.T) {
	app := setupTestApp(t, "")

	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestContextPathFromEnv(t *testing.T) {
	app := setupTestApp(t, "testsvc")

	req, _ := http.NewRequest(http.MethodGet, "/testsvc/ping", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestContextPathFallback(t *testing.T) {
	app := setupTestApp(t, "")

	req, _ := http.NewRequest(
		http.MethodGet,
		"/"+base.ServiceName+"/ping",
		nil,
	)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProtectedRoute_WithoutAPIKey(t *testing.T) {
	app := setupTestApp(t, "")

	req, _ := http.NewRequest(
		http.MethodGet,
		"/"+base.ServiceName+"/api/v1/news/search?q=test",
		nil,
	)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAppStop(t *testing.T) {
	app := New()
	err := app.Stop()
	assert.NoError(t, err)
}
