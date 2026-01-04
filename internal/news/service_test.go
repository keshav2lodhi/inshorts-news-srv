package news

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"inshorts.com/inshorts-news-srv/internal/llm"
)

type mockRepo struct {
	searchFn   func(context.Context, string, int, int) (*ResponseData, error)
	nearbyFn   func(context.Context, float64, float64, int64, int, int) (*ResponseData, error)
	categoryFn func(context.Context, string, int, int) (*ResponseData, error)
	sourceFn   func(context.Context, string, int, int) (*ResponseData, error)
	scoreFn    func(context.Context, float64, int, int) (*ResponseData, error)
}

func (m *mockRepo) Search(ctx context.Context, q string, from, size int) (*ResponseData, error) {
	return m.searchFn(ctx, q, from, size)
}
func (m *mockRepo) Nearby(ctx context.Context, lat, lon float64, radius int64, from, size int) (*ResponseData, error) {
	return m.nearbyFn(ctx, lat, lon, radius, from, size)
}
func (m *mockRepo) ByCategory(ctx context.Context, cat string, from, size int) (*ResponseData, error) {
	return m.categoryFn(ctx, cat, from, size)
}
func (m *mockRepo) BySource(ctx context.Context, src string, from, size int) (*ResponseData, error) {
	return m.sourceFn(ctx, src, from, size)
}
func (m *mockRepo) ByScore(ctx context.Context, score float64, from, size int) (*ResponseData, error) {
	return m.scoreFn(ctx, score, from, size)
}

type mockLLM struct {
	calls int32
}

func (m *mockLLM) Analyze(ctx context.Context, q string) (*llm.Result, error) {
	return nil, nil
}

func (m *mockLLM) Summarize(ctx context.Context, text string) (string, error) {
	atomic.AddInt32(&m.calls, 1)
	return "summary: " + text, nil
}

func sampleResponse() *ResponseData {
	return &ResponseData{
		Article: []Article{
			{ID: "1", Description: "first article"},
			{ID: "2", Description: "second article"},
		},
	}
}

func TestService_Search_Success(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		searchFn: func(ctx context.Context, q string, from, size int) (*ResponseData, error) {
			return sampleResponse(), nil
		},
	}

	llmMock := &mockLLM{}
	svc := NewService(repo, llmMock)

	resp, err := svc.Search(ctx, "news", 0, 10)

	assert.NoError(t, err)
	assert.Len(t, resp.Article, 2)
	assert.Equal(t, "summary: first article", resp.Article[0].LLMSummary)
	assert.Equal(t, int32(2), atomic.LoadInt32(&llmMock.calls))
}

func TestService_Search_RepoError(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		searchFn: func(ctx context.Context, q string, from, size int) (*ResponseData, error) {
			return nil, errors.New("repo failure")
		},
	}

	svc := NewService(repo, &mockLLM{})

	resp, err := svc.Search(ctx, "news", 0, 10)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestService_Enrich_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	llmMock := &mockLLM{}

	svc := &Service{
		repo: nil,
		llm:  llmMock,
	}

	data := sampleResponse()
	svc.enrichWithLLMSummary(ctx, data)

	assert.Equal(t, "", data.Article[0].LLMSummary)
	assert.Equal(t, int32(0), atomic.LoadInt32(&llmMock.calls))
}

func TestService_OtherMethods(t *testing.T) {
	tests := []struct {
		name string
		call func(*Service) (*ResponseData, error)
	}{
		{
			name: "Nearby",
			call: func(s *Service) (*ResponseData, error) {
				return s.Nearby(context.Background(), 1, 1, 5, 0, 10)
			},
		},
		{
			name: "ByCategory",
			call: func(s *Service) (*ResponseData, error) {
				return s.ByCategory(context.Background(), "tech", 0, 10)
			},
		},
		{
			name: "BySource",
			call: func(s *Service) (*ResponseData, error) {
				return s.BySource(context.Background(), "reuters", 0, 10)
			},
		},
		{
			name: "ByScore",
			call: func(s *Service) (*ResponseData, error) {
				return s.ByScore(context.Background(), 10, 0, 10)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llmMock := &mockLLM{}
			repo := &mockRepo{
				nearbyFn: func(context.Context, float64, float64, int64, int, int) (*ResponseData, error) {
					return sampleResponse(), nil
				},
				categoryFn: func(context.Context, string, int, int) (*ResponseData, error) { return sampleResponse(), nil },
				sourceFn:   func(context.Context, string, int, int) (*ResponseData, error) { return sampleResponse(), nil },
				scoreFn:    func(context.Context, float64, int, int) (*ResponseData, error) { return sampleResponse(), nil },
			}

			svc := NewService(repo, llmMock)
			resp, err := tt.call(svc)

			assert.NoError(t, err)
			assert.Equal(t, int32(2), atomic.LoadInt32(&llmMock.calls))
			assert.NotEmpty(t, resp.Article[0].LLMSummary)
		})
	}
}
