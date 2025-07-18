package feed

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

	StoredLastSavedItem time.Time
}

type FeedItem struct {
	ID          string                 `json:"id"`
	FeedTitle   string                 `json:"feed_title"`
	Title       string                 `json:"title"`
	Link        string                 `json:"link"`
	ImageURL    string                 `json:"image_url"`
	Description string                 `json:"description"`
	PublishedAt *time.Time             `json:"published_at"`
	UpdatedAt   *time.Time             `json:"updated_at"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

func (fi *FeedItem) GetMetadataJson() (string, error) {
	if fi.Metadata == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(fi.Metadata)
	return string(bytes), err
}

func (fi *FeedItem) GetTagsJson() (string, error) {
	if len(fi.Tags) == 0 {
		return "[]", nil
	}
	bytes, err := json.Marshal(fi.Tags)
	return string(bytes), err
}

func NewFeedItem(feedTitle, title, link, imageURL, description string, publishedAt *time.Time, updatedAt *time.Time, tags []string) FeedItem {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s__%s__%s__%s", title, link, description, imageURL)))
	itemID := fmt.Sprintf("%x", h.Sum(nil))
	return FeedItem{
		ID:          itemID,
		FeedTitle:   feedTitle,
		Title:       title,
		Link:        link,
		ImageURL:    imageURL,
		Description: description,
		PublishedAt: publishedAt,
		UpdatedAt:   updatedAt,
		Tags:        tags,
		Metadata:    nil,
	}
}

type FeedConfig struct {
	Name            string   `json:"name" yaml:"name"`
	URL             string   `json:"url" yaml:"url"`
	Key             string   `json:"key" yaml:"key"`
	DescriptionType string   `json:"description_type" yaml:"description_type"`
	Tags            []string `json:"tags" yaml:"tags"`
}
