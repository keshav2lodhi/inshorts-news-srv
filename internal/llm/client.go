package llm

import "context"

type Client interface {
	Analyze(ctx context.Context, query string) (*Result, error)
	Summarize(ctx context.Context, text string) (string, error)
}
