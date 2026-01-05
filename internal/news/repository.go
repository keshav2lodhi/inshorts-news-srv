package news

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
	"github.com/rs/zerolog/log"
	"inshorts.com/inshorts-news-srv/internal/base"
)

type RepositoryAPI interface {
	Search(ctx context.Context, q string, from, size int) (*ResponseData, error)
	Nearby(ctx context.Context, lat, lon float64, radiusKm int64, from, size int) (*ResponseData, error)
	ByCategory(ctx context.Context, cat string, from, size int) (*ResponseData, error)
	BySource(ctx context.Context, src string, from, size int) (*ResponseData, error)
	ByScore(ctx context.Context, minScore float64, from, size int) (*ResponseData, error)
}

type Repository struct {
	es *elasticsearch.Client
}

type ESSearchResponse struct {
	Took int64 `json:"took"`
	Hits struct {
		Total struct {
			Value    int64  `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			Score  float64 `json:"_score,omitempty"`
			Source Article `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type ResponseData struct {
	Took     int64     `json:"took_ms"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pagesize"`
	Count    int       `json:"count"`
	Article  []Article `json:"articles"`
}

func NewRepository(es *elasticsearch.Client) *Repository {
	return &Repository{es: es}
}

func (r *Repository) Search(ctx context.Context, queryParam string, from, size int) (*ResponseData, error) {
	// ES DLS query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"function_score": map[string]interface{}{
				"query": map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query": queryParam,
						"type":  "best_fields",
						"fields": []string{
							"title^3",
							"description^2",
							"llm_summary",
						},
						"operator": "and",
					},
				},
				"functions": []map[string]interface{}{
					{
						"field_value_factor": map[string]interface{}{
							"field":   "relevance_score",
							"factor":  1.2,
							"missing": 0.1,
						},
					},
				},
				"boost_mode": "sum",
				"score_mode": "sum",
			},
		},
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		log.Info().Caller().Msgf("error marshalling search query JSON: (%v)", err)
	}
	log.Info().Caller().Msgf("query for search is: (%s)", string(jsonData))

	data, err := r.executeSearch(ctx, query, from, size)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Repository) Nearby(ctx context.Context, lat, lon float64, radiusKm int64, from, size int) (*ResponseData, error) {
	// ES DLS query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{
						"geo_distance": map[string]interface{}{
							"distance": fmt.Sprintf("%dkm", radiusKm),
							"location": map[string]interface{}{
								"lat": lat,
								"lon": lon,
							},
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"_geo_distance": map[string]interface{}{
					"location": map[string]interface{}{
						"lat": lat,
						"lon": lon,
					},
					"order": "asc",
					"unit":  "km",
				},
			},
			{
				"publication_date": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		log.Info().Caller().Msgf("error marshalling nearby query JSON: (%v)", err)
	}
	log.Info().Caller().Msgf("query for nearby is: (%s)", string(jsonData))

	data, err := r.executeSearch(ctx, query, from, size)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Repository) ByCategory(ctx context.Context, cat string, from, size int) (*ResponseData, error) {
	// ES DLS query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"category": map[string]interface{}{
					"query":    cat,
					"operator": "and",
				},
			},
		},
		"sort": map[string]interface{}{
			"publication_date": "desc",
		},
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		log.Info().Caller().Msgf("error marshalling category query JSON: (%v)", err)
	}
	log.Info().Caller().Msgf("query for category is: (%s)", string(jsonData))

	data, err := r.executeSearch(ctx, query, from, size)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Repository) BySource(ctx context.Context, src string, from, size int) (*ResponseData, error) {
	// ES DLS query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"source_name": map[string]interface{}{
					"query":    src,
					"operator": "and",
				},
			},
		},
		"sort": map[string]interface{}{
			"publication_date": "desc",
		},
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		log.Info().Caller().Msgf("error marshalling source query JSON: (%v)", err)
	}
	log.Info().Caller().Msgf("query for source is: (%s)", string(jsonData))

	data, err := r.executeSearch(ctx, query, from, size)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Repository) ByScore(ctx context.Context, minScore float64, from, size int) (*ResponseData, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"relevance_score": map[string]interface{}{
					"gte": minScore,
				},
			},
		},
		"sort": []map[string]interface{}{
			{"relevance_score": "desc"},
			{"publication_date": "desc"},
		},
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		log.Info().Caller().Msgf("error marshalling score query JSON: (%v)", err)
	}
	log.Info().Caller().Msgf("query for score is: (%s)", string(jsonData))

	data, err := r.executeSearch(ctx, query, from, size)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseSearchResponse(res *esapi.Response) (
	took int64,
	total int64,
	articles []Article,
	err error,
) {
	// Close the response body
	defer res.Body.Close()

	if res.IsError() {
		return 0, 0, nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	var esResp ESSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return 0, 0, nil, err
	}

	articles = make([]Article, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		article := hit.Source
		articles = append(articles, article)
	}

	return esResp.Took, esResp.Hits.Total.Value, articles, nil
}

func (r *Repository) LoadAllArticles(ctx context.Context) (map[string]Article, error) {

	query := map[string]interface{}{
		"size": base.MaxArticlesForTrending, // enough for assignment
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex(base.ESIndex),
		r.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("es error: %s", res.String())
	}

	var esResp struct {
		Hits struct {
			Hits []struct {
				Source Article `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	articles := make(map[string]Article)

	for _, hit := range esResp.Hits.Hits {
		a := hit.Source
		articles[a.ID] = a
	}

	return articles, nil
}

func (r *Repository) executeSearch(ctx context.Context, query map[string]interface{}, from, size int) (*ResponseData, error) {
	// Send search request to es
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex(base.ESIndex),
		r.es.Search.WithBody(bytes.NewReader(body)),
		r.es.Search.WithTrackTotalHits(true),
		r.es.Search.WithFrom(from),
		r.es.Search.WithSize(size),
	)
	if err != nil {
		return nil, err
	}

	took, total, articles, err := parseSearchResponse(res)
	if err != nil {
		return nil, err
	}

	return BuildResponseData(took, total, from, size, articles), nil
}

func BuildResponseData(took int64, total int64, from int, size int, data []Article) *ResponseData {
	page := 1
	if from == 0 {
		page = 1
	} else {
		page = (from / size) + 1
	}

	return &ResponseData{
		Took:     took,
		Total:    total,
		Page:     page,
		PageSize: size,
		Count:    len(data),
		Article:  data,
	}
}
