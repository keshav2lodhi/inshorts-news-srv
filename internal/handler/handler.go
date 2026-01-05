package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"inshorts.com/inshorts-news-srv/internal/base"
	"inshorts.com/inshorts-news-srv/internal/trending"

	"inshorts.com/inshorts-news-srv/internal/news"
)

type Handler struct {
	service  news.ServiceAPI
	trending *trending.TrendingService
	articles map[string]news.Article
}

func NewHandler(es *elasticsearch.Client, newsService *news.Service, trendingSvc *trending.TrendingService, articles map[string]news.Article) *Handler {
	return &Handler{
		service:  newsService,
		trending: trendingSvc,
		articles: articles,
	}
}

// NewHandlerWithDeps is ONLY for tests
func NewHandlerWithDeps(
	service news.ServiceAPI,
	trendingSvc *trending.TrendingService,
	articles map[string]news.Article,
) *Handler {
	return &Handler{
		service:  service,
		trending: trendingSvc,
		articles: articles,
	}
}

func (h *Handler) Search(c *fiber.Ctx) error {
	r := NewAPIResponder[*news.ResponseData](c)
	query := c.Query("q")
	if query == "" {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, "query parameter 'q' is required")
	}

	// Optional user location
	var latPtr, lonPtr *float64
	if lat := c.Query("lat"); lat != "" {
		if v, err := strconv.ParseFloat(lat, 64); err == nil {
			latPtr = &v
		}
	}
	if lon := c.Query("lon"); lon != "" {
		if v, err := strconv.ParseFloat(lon, 64); err == nil {
			lonPtr = &v
		}
	}

	from, size, err := parsePagination(c)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, err.Error())
	}
	result, err := h.service.Search(c.Context(), query, latPtr, lonPtr, from, size)
	if err != nil {
		return r.RespondWithError(fiber.StatusInternalServerError, base.CodeInternalError, err.Error())
	}

	return r.RespondWithSuccess(http.StatusOK, &result)
}

func (h *Handler) Nearby(c *fiber.Ctx) error {
	r := NewAPIResponder[*news.ResponseData](c)
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, "invalid latitude")
	}

	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, "invalid longitude")
	}

	radiusKm, err := strconv.ParseInt(c.Query("radiusKm"), 10, 64)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, "invalid radiusKm")
	}

	from, size, err := parsePagination(c)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, err.Error())
	}

	result, err := h.service.Nearby(c.Context(), lat, lon, radiusKm, from, size)
	if err != nil {
		return r.RespondWithError(fiber.StatusInternalServerError, base.CodeInternalError, err.Error())
	}

	return r.RespondWithSuccess(http.StatusOK, &result)
}

func (h *Handler) Category(c *fiber.Ctx) error {
	r := NewAPIResponder[*news.ResponseData](c)

	cat := c.Query("category")
	if cat == "" {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, "category is required")
	}

	from, size, err := parsePagination(c)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, err.Error())
	}

	result, err := h.service.ByCategory(c.Context(), cat, from, size)
	if err != nil {
		return r.RespondWithError(fiber.StatusInternalServerError, base.CodeInternalError, err.Error())
	}

	return r.RespondWithSuccess(http.StatusOK, &result)
}

func (h *Handler) Source(c *fiber.Ctx) error {
	r := NewAPIResponder[*news.ResponseData](c)

	source := c.Query("source")
	if source == "" {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, "source is required")
	}

	from, size, err := parsePagination(c)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, err.Error())
	}

	result, err := h.service.BySource(c.Context(), source, from, size)
	if err != nil {
		return r.RespondWithError(fiber.StatusInternalServerError, base.CodeInternalError, err.Error())
	}

	return r.RespondWithSuccess(http.StatusOK, &result)
}

func (h *Handler) Score(c *fiber.Ctx) error {
	r := NewAPIResponder[*news.ResponseData](c)

	// default min score 0.7
	minScore := base.DefaultMinScore

	if v := c.Query("min_score"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			minScore = parsed
		}
	}

	from, size, err := parsePagination(c)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, err.Error())
	}

	result, err := h.service.ByScore(c.Context(), minScore, from, size)
	if err != nil {
		return r.RespondWithError(fiber.StatusInternalServerError, base.CodeInternalError, err.Error())
	}

	return r.RespondWithSuccess(http.StatusOK, &result)
}

func (h *Handler) Trending(c *fiber.Ctx) error {
	r := NewAPIResponder[*news.ResponseData](c)

	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, "invalid latitude")
	}

	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		return r.RespondWithError(fiber.StatusBadRequest, base.CodeInvalidData, "invalid longitude")
	}

	limit := c.QueryInt("limit", base.DefaultPageSize)

	result, err := h.trending.Trending(c.Context(), lat, lon, limit, h.articles)
	if err != nil {
		return r.RespondWithError(fiber.StatusInternalServerError, base.CodeInternalError, err.Error())
	}
	return r.RespondWithSuccess(http.StatusOK, &result)
}

func parsePagination(c *fiber.Ctx) (from, size int, err error) {
	from = c.QueryInt("from")
	if from == 0 {
		from = 0
	}
	size = c.QueryInt("size")
	if size == 0 {
		size = base.DefaultPageSize
	}

	if from+size > base.MaxWindowSize {
		err := fmt.Errorf("too large window: from + size must be <= %d", base.MaxWindowSize)
		log.Error().Caller().Msgf("invalid query param from and size is passed. Reason:(%v)", err)
		return 0, 0, err
	}
	return from, size, nil
}
