package news

import (
	"context"

	"inshorts.com/inshorts-news-srv/internal/llm"
	// "inshorts.com/inshorts-news-srv/internal/utils"
)

type Service struct {
	repo *Repository
	llm  llm.Client
}

func NewService(r *Repository, l llm.Client) *Service {
	return &Service{repo: r, llm: l}
}

func (s *Service) Search(ctx context.Context, q string, from int, size int) (*ResponseData, error) {
	result, err := s.repo.Search(ctx, q, from, size)
	if err != nil {
		return nil, err
	}
	// Enrich with LLM summary
	for i := range result.Article {
		result.Article[i].LLMSummary = s.llm.Summarize(result.Article[i].Description)
	}
	return result, nil
}

func (s *Service) Nearby(ctx context.Context, lat float64, lon float64, radiusKm int64, from int, size int) (*ResponseData, error) {
	result, err := s.repo.Nearby(ctx, lat, lon, radiusKm, from, size)
	if err != nil {
		return nil, err
	}

	// Calculate distance for ranking
	// for i := range articles {
	// 	d := utils.DistanceKm(
	// 		lat, lon,
	// 		articles[i].Latitude,
	// 		articles[i].Longitude,
	// 	)
	// 	articles[i].RelevanceScore = 1 / (1 + d) // closer = higher score
	// }

	// Sort by distance (implicitly via score)
	// RankByScore(articles)

	// if len(articles) > 5 {
	// 	return articles[:5], nil
	// }

	// Enrich with LLM summary
	for i := range result.Article {
		result.Article[i].LLMSummary = s.llm.Summarize(result.Article[i].Description)
	}
	return result, nil
}

func (s *Service) ByCategory(ctx context.Context, cat string, from int, size int) (*ResponseData, error) {
	return s.repo.ByCategory(ctx, cat, from, size)
}

func (s *Service) BySource(ctx context.Context, src string, from int, size int) (*ResponseData, error) {
	return s.repo.BySource(ctx, src, from, size)
}

func (s *Service) ByScore(ctx context.Context, minScore float64, from int, size int) (*ResponseData, error) {
	return s.repo.ByScore(ctx, minScore, from, size)
}