package news

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockESClient(responseBody string, status int) *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Header: http.Header{
					"X-Elastic-Product": []string{"Elasticsearch"},
					"Content-Type":      []string{"application/json"},
				},
			}, nil
		}),
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return es
}

/*
-------------------------------------------------
Common ES Search Response
-------------------------------------------------
*/

const validSearchResponse = `{
	"took": 7,
	"hits": {
		"total": { "value": 2, "relation": "eq" },
		"hits": [
			{ "_score": 1.5, "_source": { "id": "a1", "title": "Article 1" } },
			{ "_score": 1.2, "_source": { "id": "a2", "title": "Article 2" } }
		]
	}
}`

/*
-------------------------------------------------
Search()
-------------------------------------------------
*/

func TestRepository_Search(t *testing.T) {
	es := mockESClient(validSearchResponse, http.StatusOK)
	repo := NewRepository(es)

	data, err := repo.Search(context.Background(), "news", 0, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), data.Total)
	assert.Equal(t, 2, data.Count)
	assert.Equal(t, 1, data.Page)
	assert.Equal(t, "a1", data.Article[0].ID)
}

/*
-------------------------------------------------
Nearby()
-------------------------------------------------
*/

func TestRepository_Nearby(t *testing.T) {
	es := mockESClient(validSearchResponse, http.StatusOK)
	repo := NewRepository(es)

	data, err := repo.Nearby(context.Background(), 12.9, 77.5, 10, 0, 5)

	assert.NoError(t, err)
	assert.Len(t, data.Article, 2)
}

/*
-------------------------------------------------
ByCategory()
-------------------------------------------------
*/

func TestRepository_ByCategory(t *testing.T) {
	es := mockESClient(validSearchResponse, http.StatusOK)
	repo := NewRepository(es)

	data, err := repo.ByCategory(context.Background(), "sports", 0, 5)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), data.Total)
}

/*
-------------------------------------------------
BySource()
-------------------------------------------------
*/

func TestRepository_BySource(t *testing.T) {
	es := mockESClient(validSearchResponse, http.StatusOK)
	repo := NewRepository(es)

	data, err := repo.BySource(context.Background(), "bbc", 0, 5)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(data.Article))
}

/*
-------------------------------------------------
ByScore()
-------------------------------------------------
*/

func TestRepository_ByScore(t *testing.T) {
	es := mockESClient(validSearchResponse, http.StatusOK)
	repo := NewRepository(es)

	data, err := repo.ByScore(context.Background(), 0.7, 0, 5)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), data.Total)
}

/*
-------------------------------------------------
LoadAllArticles()
-------------------------------------------------
*/

func TestRepository_LoadAllArticles(t *testing.T) {
	resp := `{
		"hits": {
			"hits": [
				{ "_source": { "id": "a1", "title": "Article 1" } },
				{ "_source": { "id": "a2", "title": "Article 2" } }
			]
		}
	}`

	es := mockESClient(resp, http.StatusOK)
	repo := NewRepository(es)

	articles, err := repo.LoadAllArticles(context.Background())

	assert.NoError(t, err)
	assert.Len(t, articles, 2)
	assert.Equal(t, "Article 1", articles["a1"].Title)
}

/*
-------------------------------------------------
parseSearchResponse() — Error Case
-------------------------------------------------
*/

func TestParseSearchResponse_Error(t *testing.T) {
	es := mockESClient(`{}`, http.StatusInternalServerError)

	res, _ := es.Search()
	_, _, _, err := parseSearchResponse(res)

	assert.Error(t, err)
}

/*
-------------------------------------------------
Pagination Calculation
-------------------------------------------------
*/

func TestBuildResponseData_PageCalculation(t *testing.T) {
	data := BuildResponseData(5, 100, 20, 10, []Article{{ID: "a1"}})

	assert.Equal(t, 3, data.Page) // (20/10)+1
	assert.Equal(t, 1, data.Count)
}
