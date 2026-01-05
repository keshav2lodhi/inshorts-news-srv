package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

type OpenRouterClient struct {
	apiKey string
	client *http.Client
}

func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *OpenRouterClient) Analyze(ctx context.Context, query string) (*AnalysisResult, error) {
	prompt := fmt.Sprintf(`
You are an NLP system designed specifically for a NEWS SEARCH ENGINE.

The news dataset contains articles with:
- Titles and descriptions mentioning people, organizations, places, and events
- Sources like "Times Now", "RT International", "Moneycontrol", "X"
- Categories like "world", "technology", "business", "politics"
- Geographic relevance using latitude/longitude

Your task:
1. Understand the user's intent to select the BEST news retrieval strategy.
2. Extract ONLY high-confidence entities useful for news filtering and ranking.
3. Use world knowledge ONLY when the association is obvious and unambiguous.
4. Prefer precision over recall. Do NOT hallucinate.

Intent rules:
- "nearby": if a place or proximity is mentioned
- "source": if a publisher/news outlet is mentioned
- "category": if a broad topic like technology, business, world is dominant
- "search": otherwise

Entity rules:
- People: real individuals mentioned or clearly implied
- Organizations: companies, platforms, political groups
- Locations: countries, cities, regions
- Topics: broad news domains (politics, technology, business, world, science)

Return STRICT JSON ONLY in this format:

{
  "intent": "search | nearby | category | source",
  "people": [],
  "organizations": [],
  "locations": [],
  "topics": []
}

Examples:

Query: "Elon Musk Twitter news"
Output:
{
  "intent": "search",
  "people": ["Elon Musk"],
  "organizations": ["Twitter"],
  "locations": [],
  "topics": ["technology", "business"]
}

Query: "Tesla news near South Africa"
Output:
{
  "intent": "nearby",
  "people": [],
  "organizations": ["Tesla"],
  "locations": ["South Africa"],
  "topics": ["business", "technology"]
}

Query: "World politics news from RT International"
Output:
{
  "intent": "source",
  "people": [],
  "organizations": ["RT International"],
  "locations": [],
  "topics": ["politics", "world"]
}

Query:
"%s"
`, query)

	raw, err := c.callLLM(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var parsed RawAnalysis
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	return Normalize(&parsed), nil
}

func (c *OpenRouterClient) Summarize(ctx context.Context, text string) (string, error) {
	prompt := fmt.Sprintf(
		"Summarize in one factual sentence (≤20 words): %s",
		text,
	)
	return c.callLLM(ctx, prompt)
}

func (c *OpenRouterClient) callLLM(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": "openai/gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", openRouterURL, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"openrouter error: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("openrouter returned empty choices")
	}
	return res.Choices[0].Message.Content, nil
}
