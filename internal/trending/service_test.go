package trending

// import (
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"inshorts.com/inshorts-news-srv/internal/news"
// )

// func TestTrendingScoreCalculation(t *testing.T) {
// 	svc := NewTrendingService()

// 	articles := map[string]news.Article{
// 		"a1": {
// 			ID:        "a1",
// 			Latitude:  12.97,
// 			Longitude: 77.59,
// 		},
// 	}

// 	event := UserEvent{
// 		ArticleID: "a1",
// 		EventType: EventClick,
// 		Lat:       12.97,
// 		Lon:       77.59,
// 		Timestamp: time.Now(),
// 	}

// 	svc.AddEvent(event)

// 	result := svc.getTrending(12.97, 77.59, 5, articles)

// 	assert.Len(t, result, 1)
// 	assert.Equal(t, "a1", result[0].ID)
// }

// func TestTrendingRadiusFilter(t *testing.T) {
// 	svc := NewTrendingService()

// 	articles := map[string]news.Article{
// 		"a1": {ID: "a1", Latitude: 10, Longitude: 10},
// 	}

// 	svc.AddEvent(UserEvent{
// 		ArticleID: "a1",
// 		EventType: EventView,
// 		Lat:       50,
// 		Lon:       50,
// 		Timestamp: time.Now(),
// 	})

// 	result := svc.getTrending(10, 10, 5, articles)
// 	assert.Len(t, result, 0)
// }

// func TestTrendingCache(t *testing.T) {
// 	svc := NewTrendingService()

// 	articles := map[string]news.Article{
// 		"a1": {ID: "a1", Latitude: 12, Longitude: 77},
// 	}

// 	svc.AddEvent(UserEvent{
// 		ArticleID: "a1",
// 		EventType: EventView,
// 		Lat:       12,
// 		Lon:       77,
// 		Timestamp: time.Now(),
// 	})

// 	first := svc.getTrending(12, 77, 5, articles)
// 	second := svc.getTrending(12, 77, 5, articles)

// 	assert.Equal(t, first, second)
// }
