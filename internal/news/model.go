package news

import "time"

type Article struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	URL             string    `json:"url"`
	PublicationDate time.Time `json:"publication_date"`
	SourceName      string    `json:"source_name"`
	Category        []string  `json:"category"`
	RelevanceScore  float64   `json:"relevance_score"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	LLMSummary      string    `json:"llm_summary"`
}
