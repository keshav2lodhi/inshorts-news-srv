package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/elastic/go-elasticsearch/v9"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"inshorts.com/inshorts-news-srv/internal/app"
	"inshorts.com/inshorts-news-srv/internal/base"
	"inshorts.com/inshorts-news-srv/internal/llm"

	// "inshorts.com/inshorts-news-srv/internal/configs"
	"inshorts.com/inshorts-news-srv/internal/news"
	"inshorts.com/inshorts-news-srv/internal/trending"
)

func main() {
	// Set loggers
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// Set the global time format for zerolog
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z"
	// Optional: force UTC to ensure 'Z' (Zulu time) is used instead of a numeric offset
	zerolog.TimestampFieldName = "@timestamp" // example for compatibility with some log processors
	logger := log.Logger

	// Cancel context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Validate the ENV configs in production before starting the app
	// err := configs.Initialize()
	// if err != nil {
	// 	logger.Fatal().Err(err).Msg("loading configs failed")
	// }

	// I hardcoded locally, but production reads from env/secret manager.
	username := os.Getenv(base.EnvESUserName)
	if username == "" {
		username = "elastic"
	}
	password := os.Getenv(base.EnvESPassword)
	if password == "" {
		password = "UMEFncAL6JL_kBNauzej"
	}

	// Elasticsearch config
	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200",
		},
		Username: username,
		Password: password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Elasticsearch client initialisation
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create elasticsearch client")
	}

	// Optional health check
	res, err := es.Info()
	if err != nil {
		logger.Fatal().Err(err).Msg("elasticsearch not reachable")
	}
	defer res.Body.Close()

	// Load articles once at startup
	newsRepo := news.NewRepository(es)
	articles, err := newsRepo.LoadAllArticles(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msgf("failed to load articles: %v", err)
	}

	// LLM Client (OpenRouter)
	openRouterKey := os.Getenv(base.EnvOpenRouterAPIKey)
	if openRouterKey == "" {
		openRouterKey = "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"
	}
	llmClient := llm.NewOpenRouterClient(openRouterKey)

	// Service
	newsService := news.NewService(newsRepo, llmClient)
	trendingSvc := trending.NewTrendingService(newsService)

	// Start background event stream simulation
	go trendingSvc.StartEventSimulation(ctx, articles)

	// create service
	port := os.Getenv(base.EnvPort)
	if port == "" {
		port = "3000"
	}
	server := app.New()

	// listen on external signal to shut down service gracefully
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		logger.Info().Msg("shutdown signal received")
		cancel()
		_ = server.Stop()
	}()

	logger.Info().Msg("service starting...")
	server.Start(&logger, port, es, newsService, trendingSvc, articles)
}
