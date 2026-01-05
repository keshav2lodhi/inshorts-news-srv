package llm

// Raw LLM output (matches JSON from OpenRouter)
type RawAnalysis struct {
	Intent    string   `json:"intent"`
	People    []string `json:"people"`
	Orgs      []string `json:"organizations"`
	Locations []string `json:"locations"`
	Topics    []string `json:"topics"`
}

// Application-level normalized result
type AnalysisResult struct {
	Intent   Intent
	Entities []string
	Location string
}
