package llm

import "context"


type Intent string

const (
	IntentSearch   Intent = "search"
	IntentNearby   Intent = "nearby"
	IntentCategory Intent = "category"
	IntentSource   Intent = "source"
)

type Client interface {
	Analyze(ctx context.Context, query string) (*AnalysisResult, error)
	Summarize(ctx context.Context, text string) (string, error)
}
