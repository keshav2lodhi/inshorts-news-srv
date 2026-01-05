package llm

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

func newTestClient(t *testing.T, handler func(*http.Request) (*http.Response, error)) *OpenRouterClient {
	t.Helper()

	return &OpenRouterClient{
		apiKey: "test-key",
		client: &http.Client{
			Transport: &mockRoundTripper{fn: handler},
		},
	}
}

func TestCallLLM_Success(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Bearer test-key", req.Header.Get("Authorization"))

		respBody := `{
			"choices": [
				{
					"message": {
						"content": "Hello from LLM"
					}
				}
			]
		}`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(respBody)),
			Header:     make(http.Header),
		}, nil
	})

	out, err := client.callLLM(context.Background(), "test prompt")

	assert.NoError(t, err)
	assert.Equal(t, "Hello from LLM", out)
}

func TestCallLLM_HTTPError(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(bytes.NewBufferString(`rate limit`)),
			Header:     make(http.Header),
		}, nil
	})

	out, err := client.callLLM(context.Background(), "test")

	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "status=429")
}

func TestCallLLM_EmptyChoices(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
		respBody := `{"choices":[]}`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(respBody)),
			Header:     make(http.Header),
		}, nil
	})

	out, err := client.callLLM(context.Background(), "test")

	assert.Error(t, err)
	assert.Equal(t, "openrouter returned empty choices", err.Error())
	assert.Empty(t, out)
}

func TestSummarize_Success(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
		resp := `{
			"choices": [{
				"message": {
					"content": "Short summary."
				}
			}]
		}`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(resp)),
			Header:     make(http.Header),
		}, nil
	})

	out, err := client.Summarize(context.Background(), "Long article text")

	assert.NoError(t, err)
	assert.Equal(t, "Short summary.", out)
}

func TestAnalyze_Success(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
		resp := `{
			"choices": [{
				"message": {
					"content": "{\"intent\":\"source\",\"people\":[],\"organizations\":[\"RT International\"],\"locations\":[],\"topics\":[\"world\"]}"
				}
			}]
		}`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(resp)),
			Header:     make(http.Header),
		}, nil
	})

	result, err := client.Analyze(context.Background(), "World news from RT International")

	assert.NoError(t, err)
	assert.Equal(t, IntentSource, result.Intent)
	assert.Contains(t, result.Entities, "RT International")
}
