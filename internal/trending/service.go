package trending

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"inshorts.com/inshorts-news-srv/internal/base"
	"inshorts.com/inshorts-news-srv/internal/news"
	"inshorts.com/inshorts-news-srv/internal/utils"
)

type TrendingService struct {
	mu      sync.RWMutex
	events  []UserEvent
	cache   map[string]CachedTrending
	newsSvc *news.Service
}

func NewTrendingService(newsSvc *news.Service) *TrendingService {
	return &TrendingService{
		events:  make([]UserEvent, 0),
		cache:   make(map[string]CachedTrending),
		newsSvc: newsSvc,
	}
}

func (s *TrendingService) Trending(ctx context.Context, lat float64, lon float64, limit int, articles map[string]news.Article) (*news.ResponseData, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if limit <= 0 {
		limit = base.DefaultLimit
	}
	if limit > base.MaxLimit {
		limit = base.MaxLimit
	}

	start := time.Now()
	trending := s.getTrending(lat, lon, limit, articles)
	elapsed := time.Since(start)
	tookMs := elapsed.Microseconds() / 1000
	if tookMs == 0 {
		tookMs = 1
	}
	data := news.BuildResponseData(tookMs, int64(len(trending)), 0, limit, trending)

	// Enrich with LLM summary
	s.newsSvc.EnrichWithLLMSummary(ctx, data)
	return data, nil
}

func (s *TrendingService) AddEvent(e UserEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
	log.Info().Caller().Msgf("adding article: (%v) with more: (%v) from this location lat: (%v) and lon: (%v) at this time: (%v) \n", e.ArticleID, e.EventType, e.Lat, e.Lon, e.Timestamp)
}

func (s *TrendingService) getTrending(lat, lon float64, limit int, articles map[string]news.Article) []news.Article {
	cellKey := fmt.Sprintf("%s:%d", utils.GeoCell(lat, lon), limit)

	// Cache lookup
	s.mu.RLock()
	cached, ok := s.cache[cellKey]
	s.mu.RUnlock()

	if ok && time.Now().Before(cached.Expiry) {
		return cached.Articles
	}

	// Safe read of events
	s.mu.RLock()
	events := make([]UserEvent, len(s.events))
	copy(events, s.events)
	s.mu.RUnlock()

	scoreMap := make(map[string]float64)

	for _, e := range events {
		_, ok := articles[e.ArticleID]
		if !ok {
			continue
		}

		dist := utils.DistanceKm(lat, lon, e.Lat, e.Lon)
		// radius filter
		if dist > base.EventRadiusKm {
			continue
		}
		// Trending Score = Σ (event_weight × recency_weight × geo_weight)
		score := eventWeights[e.EventType] * recencyWeight(e.Timestamp) * (1 / (1 + dist))

		if log.Debug().Enabled() {
			log.Debug().Caller().Msgf("trending article: (%v) with more: (%v) which qualify the distance: (%v) and having trending score: (%v) from this location lat: (%v) and lon: (%v) at this time: (%v) \n", e.ArticleID, e.EventType, dist, score, e.Lat, e.Lon, e.Timestamp)
		}
		scoreMap[e.ArticleID] += score
	}

	type scoredArticle struct {
		Article news.Article
		Score   float64
	}

	scored := make([]scoredArticle, 0, len(scoreMap))
	for id, score := range scoreMap {
		scored = append(scored, scoredArticle{
			Article: articles[id],
			Score:   score,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	result := make([]news.Article, 0, limit)
	for i := 0; i < len(scored) && i < limit; i++ {
		log.Debug().Caller().Msgf("this article: (%v) is getting added in trending result", scored[i].Article.ID)
		result = append(result, scored[i].Article)
	}

	// Cache result
	s.mu.Lock()
	s.cache[cellKey] = CachedTrending{
		Articles: result,
		Expiry:   time.Now().Add(base.CacheTTL),
	}
	s.mu.Unlock()

	return result
}

func recencyWeight(eventTime time.Time) float64 {
	hoursAgo := time.Since(eventTime).Hours()
	return math.Exp(-hoursAgo / base.RecencyHalfLifeH) // 24h half-life
}

func (s *TrendingService) StartEventSimulation(
	ctx context.Context,
	articles map[string]news.Article,
) {
	if len(articles) == 0 {
		log.Warn().Msg("no articles available for event simulation")
		return
	}

	// Seed ONCE per process
	rand.Seed(time.Now().UnixNano())

	// Convert map - slice ONCE
	articleList := make([]news.Article, 0, len(articles))
	for _, a := range articles {
		articleList = append(articleList, a)
	}

	eventTicker := time.NewTicker(base.EventInterval)
	pruneTicker := time.NewTicker(base.PruneInterval)

	defer eventTicker.Stop()
	defer pruneTicker.Stop()

	for {
		select {

		//graceful shutdown
		case <-ctx.Done():
			log.Info().Msg("stopping trending event simulation")
			return

		case <-eventTicker.C:
			func() {
				// panic safety
				defer func() {
					if r := recover(); r != nil {
						log.Error().Msgf("panic in event simulation: %v", r)
					}
				}()

				a := articleList[rand.Intn(len(articleList))]

				// explicit weighted events
				eventType := EventView
				r := rand.Float64()
				if r < 0.2 {
					eventType = EventClick
				}

				// bounded geo
				latOffset := (rand.Float64() - 0.5) * 0.05
				lonOffset := (rand.Float64() - 0.5) * 0.05

				event := UserEvent{
					ArticleID: a.ID,
					EventType: eventType,
					Lat:       a.Latitude + latOffset,
					Lon:       a.Longitude + lonOffset,
					Timestamp: time.Now().UTC(),
				}

				s.AddEvent(event)
			}()

		case <-pruneTicker.C:
			s.pruneEvents(base.MaxEventAge)
		}
	}
}

// Event lifecycle
func (s *TrendingService) pruneEvents(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	s.mu.Lock()
	defer s.mu.Unlock()

	n := 0
	for _, e := range s.events {
		if e.Timestamp.After(cutoff) {
			s.events[n] = e
			n++
		}
	}
	s.events = s.events[:n]
}
