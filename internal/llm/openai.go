package llm

import (
	"context"
	"errors"
	"strings"
	"time"
)

// type Client interface {
// 	Analyze(query string) (*Result, error)
// 	Summarize(text string) string
// }

// type Client interface {
// 	Analyze(ctx context.Context, query string) (*Result, error)
// 	Summarize(ctx context.Context, text string) (string, error)
// }

// type Result struct {
// 	Intent   string   `json:"intent"`
// 	Entities []string `json:"entities"`
// }

type MockClient struct {
	maxSummaryLen int
}

func NewMockClient() *MockClient {
	return &MockClient{
		maxSummaryLen: 120,
	}
}

// func (m *MockClient) Analyze(q string) (*Result, error) {
// 	q = strings.ToLower(q)

// 	switch {
// 	case strings.Contains(q, "near"):
// 		return &Result{Intent: "nearby"}, nil
// 	case strings.Contains(q, "technology"):
// 		return &Result{Intent: "category"}, nil
// 	case strings.Contains(q, "reuters"):
// 		return &Result{Intent: "source"}, nil
// 	default:
// 		return &Result{Intent: "search"}, nil
// 	}
// }

/*
Analyze extracts:
- Intent
- Entities (simple keyword-based simulation)
*/
func (m *MockClient) Analyze(ctx context.Context, query string) (*Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	q := strings.ToLower(query)

	result := &Result{
		Intent:   "search",
		Entities: []string{},
	}

	switch {
	case strings.Contains(q, "near"):
		result.Intent = "nearby"
	case strings.Contains(q, "technology"):
		result.Intent = "category"
	case strings.Contains(q, "reuters"), strings.Contains(q, "times"):
		result.Intent = "source"
	}

	// very simple entity extraction (mocked)
	words := strings.Fields(query)
	for _, w := range words {
		if len(w) > 3 && strings.Title(w) == w {
			result.Entities = append(result.Entities, w)
		}
	}

	return result, nil
}

// func (m *MockClient) Summarize(text string) string {
// 	if len(text) > 120 {
// 		return text[:120] + "..."
// 	}
// 	return text
// }

/*
Summarize simulates LLM latency and truncation
Fail-soft behavior
*/
func (m *MockClient) Summarize(ctx context.Context, text string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(5 * time.Millisecond): // simulate latency
	}

	if text == "" {
		return "", errors.New("empty text")
	}

	if len(text) > m.maxSummaryLen {
		return text[:m.maxSummaryLen] + "...", nil
	}
	return text, nil
}
