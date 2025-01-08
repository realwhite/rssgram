package feed

import (
	"time"

	"github.com/google/uuid"
)

const FeedDescriptionTypeLink string = "link"

type Feed struct {
	ID          uuid.UUID              `json:"id"`
	Config      FeedConfig             `json:"config"`
	Title       string                 `json:"title"`
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Items       []FeedItem             `json:"items"`
	Key         string                 `json:"key"`
	Metadata    map[string]interface{} `json:"metadata"`
	Tags        []string               `json:"tags"`
}

type FeedItem struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Link        string     `json:"link"`
	ImageURL    string     `json:"image_url"`
	Description string     `json:"description"`
	PublishedAt *time.Time `json:"published_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	Tags        []string   `json:"tags"`
}

type FeedConfig struct {
	Name            string   `json:"name"`
	URL             string   `json:"url"`
	Key             string   `json:"key"`
	DescriptionType string   `json:"description_type"`
	Tags            []string `json:"tags"`
}
