package router

import (
	"crypto/sha256"
	"crypto/subtle"
	"os"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/rs/zerolog/log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"

	"inshorts.com/inshorts-news-srv/internal/base"
	"inshorts.com/inshorts-news-srv/internal/handler"
	"inshorts.com/inshorts-news-srv/internal/news"
	"inshorts.com/inshorts-news-srv/internal/trending"
)

// CreateRoutes inits all routes related to the application
func CreateRoutes(router fiber.Router, es *elasticsearch.Client, newsService *news.Service, trendingSvc *trending.TrendingService, articles map[string]news.Article) {
	h := handler.NewHandler(es, newsService, trendingSvc, articles)

	router.Get("/ping", handler.Ping)

	protected := router.Group("/api/v1", keyauth.New(keyauth.Config{
		KeyLookup: "header:apikey",
		Validator: validateAPIKey,
	}))

	protected.Get("/news/search", h.Search)
	protected.Get("/news/nearby", h.Nearby)
	protected.Get("/news/category", h.Category)
	protected.Get("/news/source", h.Source)
	protected.Get("/news/score", h.Score)
	protected.Get("/news/trending", h.Trending)
}

func validateAPIKey(c *fiber.Ctx, key string) (bool, error) {
	apiKey := os.Getenv(base.EnvAPIKey)
	if apiKey == "" {
		apiKey = "abcdxyz" // I hardcoded locally, but production reads from env/secret manager.
	}
	if apiKey == "" {
		log.Error().Caller().Msgf("the apiKey is not set in env variable for this API (%s)", c.OriginalURL())
		return false, keyauth.ErrMissingOrMalformedAPIKey
	}
	hashedAPIKey := sha256.Sum256([]byte(apiKey))
	hashedKey := sha256.Sum256([]byte(key))

	if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
		return true, nil
	}
	log.Info().Caller().Msgf("apikey is invalid for url (%s)", c.OriginalURL())
	return false, keyauth.ErrMissingOrMalformedAPIKey
}
