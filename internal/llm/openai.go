package llm

import "strings"

type Client interface {
	Analyze(query string) (*Result, error)
	Summarize(text string) string
}

type Result struct {
	Intent   string
	Entities []string
}

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) Analyze(q string) (*Result, error) {
	q = strings.ToLower(q)

	switch {
	case strings.Contains(q, "near"):
		return &Result{Intent: "nearby"}, nil
	case strings.Contains(q, "technology"):
		return &Result{Intent: "category"}, nil
	case strings.Contains(q, "reuters"):
		return &Result{Intent: "source"}, nil
	default:
		return &Result{Intent: "search"}, nil
	}
}

func (m *MockClient) Summarize(text string) string {
	if len(text) > 120 {
		return text[:120] + "..."
	}
	return text
}
