package news

import (
	"context"
	"errors"
	"testing"

	"inshorts.com/inshorts-news-srv/internal/llm"
)

type mockLLM struct {
	analyzeResp llm.AnalysisResult
	analyzeErr  error

	summaryResp string
	summaryErr  error
}

func (m *mockLLM) Analyze(ctx context.Context, q string) (*llm.AnalysisResult, error) {
	return &m.analyzeResp, m.analyzeErr
}

func (m *mockLLM) Summarize(ctx context.Context, text string) (string, error) {
	return m.summaryResp, m.summaryErr
}

type mockRepo struct {
	searchFn   func(ctx context.Context, q string, from, size int) (*ResponseData, error)
	nearbyFn   func(ctx context.Context, lat, lon float64, radiusKm int64, from, size int) (*ResponseData, error)
	categoryFn func(ctx context.Context, cat string, from, size int) (*ResponseData, error)
	sourceFn   func(ctx context.Context, src string, from, size int) (*ResponseData, error)
	scoreFn    func(ctx context.Context, minScore float64, from, size int) (*ResponseData, error)
}

func (m *mockRepo) Search(ctx context.Context, q string, from, size int) (*ResponseData, error) {
	return m.searchFn(ctx, q, from, size)
}
func (m *mockRepo) Nearby(ctx context.Context, lat, lon float64, radiusKm int64, from, size int) (*ResponseData, error) {
	return m.nearbyFn(ctx, lat, lon, radiusKm, from, size)
}
func (m *mockRepo) ByCategory(ctx context.Context, cat string, from, size int) (*ResponseData, error) {
	return m.categoryFn(ctx, cat, from, size)
}
func (m *mockRepo) BySource(ctx context.Context, src string, from, size int) (*ResponseData, error) {
	return m.sourceFn(ctx, src, from, size)
}
func (m *mockRepo) ByScore(ctx context.Context, minScore float64, from, size int) (*ResponseData, error) {
	return m.scoreFn(ctx, minScore, from, size)
}

func testResponse() *ResponseData {
	return &ResponseData{
		Article: []Article{
			{
				ID:          "1",
				Title:       "Test",
				Description: "Some description",
			},
		},
	}
}

// Test: Search - Category intent routing
func TestService_Search_CategoryIntent(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		categoryFn: func(ctx context.Context, cat string, from, size int) (*ResponseData, error) {
			if cat != "technology" {
				t.Fatalf("expected category technology, got %s", cat)
			}
			return testResponse(), nil
		},
	}

	llmClient := &mockLLM{
		analyzeResp: llm.AnalysisResult{
			Intent:   llm.IntentCategory,
			Entities: []string{"technology"},
		},
		summaryResp: "LLM summary",
	}

	svc := NewService(repo, llmClient)

	result, err := svc.Search(ctx, "tech news", nil, nil, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Article[0].LLMSummary == "" {
		t.Fatal("expected llm summary to be set")
	}
}

// Test: Search - Source intent routing
func TestService_Search_SourceIntent(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		sourceFn: func(ctx context.Context, src string, from, size int) (*ResponseData, error) {
			if src != "bbc" {
				t.Fatalf("expected source bbc, got %s", src)
			}
			return testResponse(), nil
		},
	}

	llmClient := &mockLLM{
		analyzeResp: llm.AnalysisResult{
			Intent:   llm.IntentSource,
			Entities: []string{"bbc"},
		},
		summaryResp: "summary",
	}

	svc := NewService(repo, llmClient)

	_, err := svc.Search(ctx, "bbc news", nil, nil, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test: Nearby intent fallback (missing lat/lon)
func TestService_Search_NearbyFallback(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		searchFn: func(ctx context.Context, q string, from, size int) (*ResponseData, error) {
			return testResponse(), nil
		},
	}

	llmClient := &mockLLM{
		analyzeResp: llm.AnalysisResult{
			Intent: llm.IntentNearby,
		},
		summaryResp: "summary",
	}

	svc := NewService(repo, llmClient)

	_, err := svc.Search(ctx, "near me", nil, nil, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test: EnrichWithLLMSummary fail-soft
func TestService_EnrichWithLLMSummary_FailSoft(t *testing.T) {
	ctx := context.Background()

	llmClient := &mockLLM{
		summaryErr: errors.New("llm failure"),
	}

	svc := NewService(nil, llmClient)

	data := testResponse()

	svc.EnrichWithLLMSummary(ctx, data)

	if data.Article[0].LLMSummary != "" {
		t.Fatal("expected empty summary on llm error")
	}
}

// Test: ByScore happy path
func TestService_ByScore(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		scoreFn: func(ctx context.Context, min float64, from, size int) (*ResponseData, error) {
			if min != 0.8 {
				t.Fatalf("expected min score 0.8")
			}
			return testResponse(), nil
		},
	}

	llmClient := &mockLLM{
		summaryResp: "summary",
	}

	svc := NewService(repo, llmClient)

	result, err := svc.ByScore(ctx, 0.8, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Article[0].LLMSummary == "" {
		t.Fatal("expected summary")
	}
}
