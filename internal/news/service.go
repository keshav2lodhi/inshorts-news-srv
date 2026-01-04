package news

import (
	"context"
	"sync"

	"inshorts.com/inshorts-news-srv/internal/llm"
)

type ServiceAPI interface {
	Search(ctx context.Context, q string, from, size int) (*ResponseData, error)
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

func (s *Service) Search(ctx context.Context, q string, from int, size int) (*ResponseData, error) {
	result, err := s.repo.Search(ctx, q, from, size)
	if err != nil {
		return nil, err
	}
	// Enrich with LLM summary
	s.enrichWithLLMSummary(ctx, result)
	return result, nil
}

func (s *Service) Nearby(ctx context.Context, lat float64, lon float64, radiusKm int64, from int, size int) (*ResponseData, error) {
	result, err := s.repo.Nearby(ctx, lat, lon, radiusKm, from, size)
	if err != nil {
		return nil, err
	}

	// Enrich with LLM summary
	s.enrichWithLLMSummary(ctx, result)
	return result, nil
}

func (s *Service) ByCategory(ctx context.Context, cat string, from int, size int) (*ResponseData, error) {
	result, err := s.repo.ByCategory(ctx, cat, from, size)
	if err != nil {
		return nil, err
	}

	// Enrich with LLM summary
	s.enrichWithLLMSummary(ctx, result)
	return result, nil
}

func (s *Service) BySource(ctx context.Context, src string, from int, size int) (*ResponseData, error) {
	result, err := s.repo.BySource(ctx, src, from, size)
	if err != nil {
		return nil, err
	}

	// Enrich with LLM summary
	s.enrichWithLLMSummary(ctx, result)
	return result, nil
}

func (s *Service) ByScore(ctx context.Context, minScore float64, from int, size int) (*ResponseData, error) {
	result, err := s.repo.ByScore(ctx, minScore, from, size)
	if err != nil {
		return nil, err
	}

	// Enrich with LLM summary
	s.enrichWithLLMSummary(ctx, result)
	return result, nil
}

// enrichWithLLMSummary enriches articles using LLM.
// Fail-soft, context-aware, and parallelized.
func (s *Service) enrichWithLLMSummary(ctx context.Context, data *ResponseData) {
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
				if err == nil {
					data.Article[idx].LLMSummary = summary
				}
			}
		}()
	}

	wg.Wait()
}
