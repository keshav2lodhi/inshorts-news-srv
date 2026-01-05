package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"inshorts.com/inshorts-news-srv/internal/base"
	"inshorts.com/inshorts-news-srv/internal/news"
	"inshorts.com/inshorts-news-srv/internal/trending"
)

type mockNewsService struct {
	searchFn   func(ctx context.Context, q string, lat, lon *float64, from, size int) (*news.ResponseData, error)
	nearbyFn   func(ctx context.Context, lat, lon float64, radius int64, from, size int) (*news.ResponseData, error)
	categoryFn func(ctx context.Context, cat string, from, size int) (*news.ResponseData, error)
	sourceFn   func(ctx context.Context, src string, from, size int) (*news.ResponseData, error)
	scoreFn    func(ctx context.Context, min float64, from, size int) (*news.ResponseData, error)
}

func (m *mockNewsService) Search(ctx context.Context, q string, lat, lon *float64, from, size int) (*news.ResponseData, error) {
	return m.searchFn(ctx, q, lat, lon, from, size)
}
func (m *mockNewsService) Nearby(ctx context.Context, lat, lon float64, radius int64, from, size int) (*news.ResponseData, error) {
	return m.nearbyFn(ctx, lat, lon, radius, from, size)
}
func (m *mockNewsService) ByCategory(ctx context.Context, cat string, from, size int) (*news.ResponseData, error) {
	return m.categoryFn(ctx, cat, from, size)
}
func (m *mockNewsService) BySource(ctx context.Context, src string, from, size int) (*news.ResponseData, error) {
	return m.sourceFn(ctx, src, from, size)
}
func (m *mockNewsService) ByScore(ctx context.Context, min float64, from, size int) (*news.ResponseData, error) {
	return m.scoreFn(ctx, min, from, size)
}

func setupTestApp(h *Handler, route string, handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Get(route, handler)
	return app
}

func successResponse() *news.ResponseData {
	return &news.ResponseData{
		Took:     10,
		Total:    1,
		Page:     1,
		PageSize: 10,
		Count:    1,
		Article:  []news.Article{{ID: "1", Title: "test"}},
	}
}

// Missing query - 400
func TestSearch_MissingQuery(t *testing.T) {
	h := NewHandlerWithDeps(&mockNewsService{}, nil, nil)

	app := setupTestApp(h, "/search", h.Search)
	req := httptest.NewRequest(http.MethodGet, "/search", nil)

	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Successful search - 200
func TestSearch_Success(t *testing.T) {
	mockSvc := &mockNewsService{
		searchFn: func(ctx context.Context, q string, lat, lon *float64, from, size int) (*news.ResponseData, error) {
			assert.Equal(t, "elon musk", q)
			return successResponse(), nil
		},
	}

	h := NewHandlerWithDeps(mockSvc, nil, nil)
	app := setupTestApp(h, "/search", h.Search)

	req := httptest.NewRequest(http.MethodGet, "/search?q=elon+musk", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	assert.True(t, body["success"].(bool))
}

// Service error - 500
func TestSearch_ServiceError(t *testing.T) {
	mockSvc := &mockNewsService{
		searchFn: func(ctx context.Context, q string, lat, lon *float64, from, size int) (*news.ResponseData, error) {
			return nil, errors.New("boom")
		},
	}

	h := NewHandlerWithDeps(mockSvc, nil, nil)
	app := setupTestApp(h, "/search", h.Search)

	req := httptest.NewRequest(http.MethodGet, "/search?q=test", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// Category Handler Tests
func TestCategory_MissingParam(t *testing.T) {
	h := NewHandlerWithDeps(&mockNewsService{}, nil, nil)
	app := setupTestApp(h, "/category", h.Category)

	req := httptest.NewRequest(http.MethodGet, "/category", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCategory_Success(t *testing.T) {
	mockSvc := &mockNewsService{
		categoryFn: func(ctx context.Context, cat string, from, size int) (*news.ResponseData, error) {
			assert.Equal(t, "technology", cat)
			return successResponse(), nil
		},
	}

	h := NewHandlerWithDeps(mockSvc, nil, nil)
	app := setupTestApp(h, "/category", h.Category)

	req := httptest.NewRequest(http.MethodGet, "/category?category=technology", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Source Handler Test
func TestSource_Success(t *testing.T) {
	mockSvc := &mockNewsService{
		sourceFn: func(ctx context.Context, src string, from, size int) (*news.ResponseData, error) {
			assert.Equal(t, "RT International", src)
			return successResponse(), nil
		},
	}

	h := NewHandlerWithDeps(mockSvc, nil, nil)
	app := setupTestApp(h, "/source", h.Source)

	req := httptest.NewRequest(http.MethodGet, "/source?source=RT+International", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Score Handler Test
func TestScore_DefaultMinScore(t *testing.T) {
	mockSvc := &mockNewsService{
		scoreFn: func(ctx context.Context, min float64, from, size int) (*news.ResponseData, error) {
			assert.Equal(t, base.DefaultMinScore, min)
			return successResponse(), nil
		},
	}

	h := NewHandlerWithDeps(mockSvc, nil, nil)
	app := setupTestApp(h, "/score", h.Score)

	req := httptest.NewRequest(http.MethodGet, "/score", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Trending Handler Test
func TestTrending_InvalidLat(t *testing.T) {
	h := NewHandlerWithDeps(nil, &trending.TrendingService{}, nil)
	app := setupTestApp(h, "/trending", h.Trending)

	req := httptest.NewRequest(http.MethodGet, "/trending?lat=abc&lon=12", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
