package router_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/stretchr/testify/assert"

	"inshorts.com/inshorts-news-srv/internal/base"
	"inshorts.com/inshorts-news-srv/internal/handler"
	"inshorts.com/inshorts-news-srv/internal/news"
	"inshorts.com/inshorts-news-srv/internal/trending"
)

/* -------------------- Mocks -------------------- */

type mockNewsService struct{}

func (m *mockNewsService) Search(ctx context.Context, q string, lat, lon *float64, from, size int) (*news.ResponseData, error) {
	return &news.ResponseData{}, nil
}
func (m *mockNewsService) Nearby(ctx context.Context, lat, lon float64, radius int64, from, size int) (*news.ResponseData, error) {
	return &news.ResponseData{}, nil
}
func (m *mockNewsService) ByCategory(ctx context.Context, cat string, from, size int) (*news.ResponseData, error) {
	return &news.ResponseData{}, nil
}
func (m *mockNewsService) BySource(ctx context.Context, src string, from, size int) (*news.ResponseData, error) {
	return &news.ResponseData{}, nil
}
func (m *mockNewsService) ByScore(ctx context.Context, minScore float64, from, size int) (*news.ResponseData, error) {
	return &news.ResponseData{}, nil
}

/* -------------------- Helpers -------------------- */

func setupApp() *fiber.App {
	app := fiber.New()

	// Set API key
	os.Setenv(base.EnvAPIKey, "test-key")

	// ---- Mock service ----
	mockSvc := &mockNewsService{}

	// ---- Trending service (safe empty) ----
	trendingSvc := trending.NewTrendingService(nil)

	// ---- Handler with mocked deps ----
	h := handler.NewHandlerWithDeps(
		mockSvc,
		trendingSvc,
		map[string]news.Article{},
	)

	// ---- Routes ----
	app.Get("/ping", handler.Ping)

	protected := app.Group("/api/v1", keyauth.New(keyauth.Config{
		KeyLookup: "header:apikey",
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			return key == "test-key", nil
		},
	}))

	protected.Get("/news/search", h.Search)
	protected.Get("/news/trending", h.Trending)

	return app
}

/* -------------------- Tests -------------------- */

func TestPingRoute_NoAuth(t *testing.T) {
	app := setupApp()

	req, _ := http.NewRequest("GET", "/ping", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProtectedRoute_WithoutAPIKey(t *testing.T) {
	app := setupApp()

	req, _ := http.NewRequest("GET", "/api/v1/news/search?q=test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProtectedRoute_WithInvalidAPIKey(t *testing.T) {
	app := setupApp()

	req, _ := http.NewRequest("GET", "/api/v1/news/search?q=test", nil)
	req.Header.Set("apikey", "wrong-key")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProtectedRoute_WithValidAPIKey(t *testing.T) {
	app := setupApp()

	req, _ := http.NewRequest("GET", "/api/v1/news/search?q=test", nil)
	req.Header.Set("apikey", "test-key")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}

func TestTrendingRoute_WithValidAPIKey(t *testing.T) {
	app := setupApp()

	req, _ := http.NewRequest(
		"GET",
		"/api/v1/news/trending?lat=12.9&lon=77.6&limit=5",
		nil,
	)
	req.Header.Set("apikey", "test-key")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}
