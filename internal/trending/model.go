package trending

import (
	"time"

	"inshorts.com/inshorts-news-srv/internal/news"
)

type UserEventType string

const (
	EventView  UserEventType = "view"
	EventClick UserEventType = "click"
)

var eventWeights = map[UserEventType]float64{
	EventView:  1.0,
	EventClick: 3.0,
}

type UserEvent struct {
	ArticleID string
	EventType UserEventType
	Lat       float64
	Lon       float64
	Timestamp time.Time
}

type CachedTrending struct {
	Articles []news.Article
	Expiry   time.Time
}
