package llm

type Result struct {
	Intent   string   `json:"intent"`
	Entities []string `json:"entities"`
}
