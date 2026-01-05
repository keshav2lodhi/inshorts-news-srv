package news

import (
	"context"
	"sync"
	"time"

	"inshorts.com/inshorts-news-srv/internal/llm"
)

type ServiceAPI interface {
	Search(ctx context.Context, q string, latPtr *float64, lonPtr *float64, from, size int) (*ResponseData, error)
	Nearby(ctx context.Context, lat, lon float64, radiusKm int64, from, size int) (*ResponseData, error)
	ByCategory(ctx context.Context, cat string, from, size int) (*ResponseData, error)
	BySource(ctx context.Context, src string, from, size int) (*ResponseData, error)
	ByScore(ctx context.Context, minScore float64, from, size int) (*ResponseData, error)
}

var _ ServiceAPI = (*Service)(nil)

type Service struct {
	repo RepositoryAPI
	llm  llm.Client
}

func NewService(r RepositoryAPI, l llm.Client) *Service {
	return &Service{repo: r, llm: l}
}

func (s *Service) Search(
	ctx context.Context,
	query string,
	userLat *float64,
	userLon *float64,
	from int,
	size int,
) (*ResponseData, error) {

	start := time.Now()

	// Ask LLM to analyze query
	analysis, err := s.llm.Analyze(ctx, query)
	if err != nil {
		return nil, err
	}
	var result *ResponseData

	// Route based on intent
	switch analysis.Intent {

	case llm.IntentNearby:
		if userLat == nil || userLon == nil {
			// fallback if location missing
			result, err = s.repo.Search(ctx, query, from, size)
			break
		}
		result, err = s.repo.Nearby(
			ctx,
			*userLat,
			*userLon,
			50, // default radius
			from,
			size,
		)

	case llm.IntentCategory:
		if len(analysis.Entities) > 0 {
			result, err = s.repo.ByCategory(ctx, analysis.Entities[0], from, size)
		} else {
			result, err = s.repo.Search(ctx, query, from, size)
		}

	case llm.IntentSource:
		if len(analysis.Entities) > 0 {
			result, err = s.repo.BySource(ctx, analysis.Entities[0], from, size)
		} else {
			result, err = s.repo.Search(ctx, query, from, size)
		}

	default:
		result, err = s.repo.Search(ctx, query, from, size)
	}

	if err != nil {
		return nil, err
	}
	// Enrich with LLM summary
	s.EnrichWithLLMSummary(ctx, result)
	result.Took = max(1, time.Since(start).Milliseconds())
	return result, nil
}

func (s *Service) Nearby(ctx context.Context, lat float64, lon float64, radiusKm int64, from int, size int) (*ResponseData, error) {
	result, err := s.repo.Nearby(ctx, lat, lon, radiusKm, from, size)
	if err != nil {
		return nil, err
	}

	// Enrich with LLM summary
	s.EnrichWithLLMSummary(ctx, result)
	return result, nil
}

func (s *Service) ByCategory(ctx context.Context, cat string, from int, size int) (*ResponseData, error) {
	result, err := s.repo.ByCategory(ctx, cat, from, size)
	if err != nil {
		return nil, err
	}

	// Enrich with LLM summary
	s.EnrichWithLLMSummary(ctx, result)
	return result, nil
}

func (s *Service) BySource(ctx context.Context, src string, from int, size int) (*ResponseData, error) {
	result, err := s.repo.BySource(ctx, src, from, size)
	if err != nil {
		return nil, err
	}

	// Enrich with LLM summary
	s.EnrichWithLLMSummary(ctx, result)
	return result, nil
}

func (s *Service) ByScore(ctx context.Context, minScore float64, from int, size int) (*ResponseData, error) {
	result, err := s.repo.ByScore(ctx, minScore, from, size)
	if err != nil {
		return nil, err
	}

	// Enrich with LLM summary
	s.EnrichWithLLMSummary(ctx, result)
	return result, nil
}

// enrichWithLLMSummary enriches articles using LLM.
// Fail-soft, context-aware, and parallelized.
func (s *Service) EnrichWithLLMSummary(ctx context.Context, data *ResponseData) {
	if data == nil || len(data.Article) == 0 {
		return
	}

	var wg sync.WaitGroup

	for i := range data.Article {
		wg.Add(1)
		idx := i

		go func() {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				summary, err := s.llm.Summarize(ctx, data.Article[idx].Description)
				if err != nil {
					// fail-soft: leave summary empty
					return
				}
				// Safe write (per-index is fine, no shared mutation)
				data.Article[idx].LLMSummary = summary
			}
		}()
	}

	wg.Wait()
}
