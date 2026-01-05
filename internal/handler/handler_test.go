package handler

// import (
// 	"context"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"inshorts.com/inshorts-news-srv/internal/news"
// )

// type mockNewsService struct {
// 	searchFn   func(context.Context, string, int, int) (*news.ResponseData, error)
// 	nearbyFn   func(context.Context, float64, float64, int64, int, int) (*news.ResponseData, error)
// 	categoryFn func(context.Context, string, int, int) (*news.ResponseData, error)
// 	sourceFn   func(context.Context, string, int, int) (*news.ResponseData, error)
// 	scoreFn    func(context.Context, float64, int, int) (*news.ResponseData, error)
// }

// func (m *mockNewsService) Search(ctx context.Context, q string, from, size int) (*news.ResponseData, error) {
// 	return m.searchFn(ctx, q, from, size)
// }
// func (m *mockNewsService) Nearby(ctx context.Context, lat, lon float64, radiusKm int64, from, size int) (*news.ResponseData, error) {
// 	return m.nearbyFn(ctx, lat, lon, radiusKm, from, size)
// }
// func (m *mockNewsService) ByCategory(ctx context.Context, cat string, from, size int) (*news.ResponseData, error) {
// 	return m.categoryFn(ctx, cat, from, size)
// }
// func (m *mockNewsService) BySource(ctx context.Context, src string, from, size int) (*news.ResponseData, error) {
// 	return m.sourceFn(ctx, src, from, size)
// }
// func (m *mockNewsService) ByScore(ctx context.Context, minScore float64, from, size int) (*news.ResponseData, error) {
// 	return m.scoreFn(ctx, minScore, from, size)
// }

// func setupTestApp(handlerFunc fiber.Handler) *fiber.App {
// 	app := fiber.New()
// 	app.Get("/", handlerFunc)
// 	return app
// }

// func TestSearch_MissingQuery(t *testing.T) {
// 	h := NewHandlerWithDeps(nil, nil, nil)

// 	app := setupTestApp(h.Search)
// 	req := httptest.NewRequest(http.MethodGet, "/?q=", nil)

// 	resp, err := app.Test(req)
// 	require.NoError(t, err)
// 	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
// }

// func TestSearch_Success(t *testing.T) {
// 	mockSvc := &mockNewsService{
// 		searchFn: func(ctx context.Context, q string, from, size int) (*news.ResponseData, error) {
// 			return &news.ResponseData{
// 				Article: []news.Article{
// 					{ID: "a1", Title: "Test"},
// 				},
// 			}, nil
// 		},
// 	}

// 	h := NewHandlerWithDeps(mockSvc, nil, nil)

// 	app := setupTestApp(h.Search)

// 	req := httptest.NewRequest(http.MethodGet, "/?q=test", nil)
// 	resp, err := app.Test(req)

// 	require.NoError(t, err)
// 	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
// }

// func TestNearby_Success(t *testing.T) {
// 	mockSvc := &mockNewsService{
// 		nearbyFn: func(ctx context.Context, lat, lon float64, radiusKm int64, from, size int) (*news.ResponseData, error) {
// 			return &news.ResponseData{
// 				Article: []news.Article{{ID: "n1"}},
// 			}, nil
// 		},
// 	}

// 	h := NewHandlerWithDeps(mockSvc, nil, nil)
// 	app := setupTestApp(h.Nearby)

// 	req := httptest.NewRequest(http.MethodGet,
// 		"/?lat=12.9&lon=77.6&radiusKm=10",
// 		nil,
// 	)

// 	resp, err := app.Test(req)
// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// }

// func TestNearby_InvalidLatitude(t *testing.T) {
// 	h := NewHandlerWithDeps(nil, nil, nil)
// 	app := setupTestApp(h.Nearby)

// 	req := httptest.NewRequest(http.MethodGet,
// 		"/?lat=abc&lon=77&radiusKm=10",
// 		nil,
// 	)

// 	resp, _ := app.Test(req)
// 	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
// }

// func TestCategory_Success(t *testing.T) {
// 	mockSvc := &mockNewsService{
// 		categoryFn: func(ctx context.Context, cat string, from, size int) (*news.ResponseData, error) {
// 			return &news.ResponseData{
// 				Article: []news.Article{{ID: "c1"}},
// 			}, nil
// 		},
// 	}

// 	h := NewHandlerWithDeps(mockSvc, nil, nil)
// 	app := setupTestApp(h.Category)

// 	req := httptest.NewRequest(http.MethodGet, "/?category=tech", nil)
// 	resp, _ := app.Test(req)

// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// }

// func TestCategory_MissingCategory(t *testing.T) {
// 	h := NewHandlerWithDeps(nil, nil, nil)

// 	app := setupTestApp(h.Category)
// 	req := httptest.NewRequest(http.MethodGet, "/", nil)

// 	resp, err := app.Test(req)
// 	require.NoError(t, err)
// 	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
// }

// func TestSource_Success(t *testing.T) {
// 	mockSvc := &mockNewsService{
// 		sourceFn: func(ctx context.Context, src string, from, size int) (*news.ResponseData, error) {
// 			return &news.ResponseData{
// 				Article: []news.Article{{ID: "s1"}},
// 			}, nil
// 		},
// 	}

// 	h := NewHandlerWithDeps(mockSvc, nil, nil)
// 	app := setupTestApp(h.Source)

// 	req := httptest.NewRequest(http.MethodGet, "/?source=reuters", nil)
// 	resp, _ := app.Test(req)

// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// }

// func TestScore_DefaultMinScore(t *testing.T) {
// 	mockSvc := &mockNewsService{
// 		scoreFn: func(ctx context.Context, minScore float64, from, size int) (*news.ResponseData, error) {
// 			assert.Equal(t, 0.7, minScore)
// 			return &news.ResponseData{}, nil
// 		},
// 	}

// 	h := NewHandlerWithDeps(mockSvc, nil, nil)
// 	app := setupTestApp(h.Score)

// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	resp, _ := app.Test(req)

// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// }

// func TestParsePagination_Overflow(t *testing.T) {
// 	app := fiber.New()
// 	app.Get("/", func(c *fiber.Ctx) error {
// 		_, _, err := parsePagination(c)
// 		assert.Error(t, err)
// 		return nil
// 	})

// 	req := httptest.NewRequest(http.MethodGet, "/?from=9999&size=9999", nil)
// 	_, err := app.Test(req)
// 	require.NoError(t, err)
// }

// type mockTrendingService struct {
// 	fn func(context.Context, float64, float64, int, map[string]news.Article) (*news.ResponseData, error)
// }

// // func (m *mockTrendingService) Trending(
// // 	ctx context.Context,
// // 	lat, lon float64,
// // 	limit int,
// // 	articles map[string]news.Article,
// // ) (*news.ResponseData, error) {
// // 	return m.fn(ctx, lat, lon, limit, articles)
// // }

// // func TestTrending_Success(t *testing.T) {
// // 	mockTrending := &mockTrendingService{
// // 		fn: func(ctx context.Context, lat, lon float64, limit int, articles map[string]news.Article) (*news.ResponseData, error) {
// // 			return &news.ResponseData{
// // 				Article: []news.Article{{ID: "t1"}},
// // 			}, nil
// // 		},
// // 	}

// // 	h := NewHandlerWithDeps(nil, mockTrending, map[string]news.Article{})
// // 	app := setupTestApp(h.Trending)

// // 	req := httptest.NewRequest(http.MethodGet, "/?lat=12&lon=77", nil)
// // 	resp, _ := app.Test(req)

// // 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// // }
