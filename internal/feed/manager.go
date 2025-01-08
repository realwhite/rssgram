package feed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"rssgram/internal"
	"rssgram/internal/storage"
	"rssgram/internal/utils"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
)

type repo interface {
	GetAllFeeds(ctx context.Context) ([]storage.StoredFeed, error)
	DeleteFeed(ctx context.Context, url string) error
	UpsertFeed(ctx context.Context, url string, lastChecked, lastPost time.Time) error
}

type Manager struct {
	repo repo
}

func (fm *Manager) EnrichFeedItems(feed *Feed) error {
	if feed.Config.DescriptionType != FeedDescriptionTypeLink {
		return nil
	}

	wg := sync.WaitGroup{}

	for i := range feed.Items {
		item := &feed.Items[i]

		wg.Add(1)
		go func() {
			defer wg.Done()
			p := internal.SiteParser{}
			siteDescription, err := p.GetDescription(item.Link)
			if err != nil {
				//	TODO: LOG
				return
			}

			if siteDescription.Description != "" {
				item.Description = siteDescription.Description
			} else if siteDescription.Title != "" {
				item.Description = siteDescription.Title
			}

			if siteDescription.Image != "" {
				item.ImageURL = siteDescription.Image
			}
		}()
	}

	wg.Wait()

	return nil
}

func (fm *Manager) GetFeed(f FeedConfig) (*Feed, error) {
	fp := gofeed.NewParser()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	respFeed, err := fp.ParseURLWithContext(f.URL, ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to parse feed URL %s: %w", f.URL, err)
	}

	var items []FeedItem

	for _, item := range respFeed.Items {

		var imageURL string
		if item.Image != nil {
			imageURL = item.Image.URL
		}

		items = append(items, FeedItem{
			ID:          uuid.UUID{},
			Title:       item.Title,
			Link:        item.Link,
			ImageURL:    imageURL,
			Description: item.Description,
			PublishedAt: item.PublishedParsed,
			UpdatedAt:   item.UpdatedParsed,
			Tags:        utils.MergeStrSlices(item.Categories, f.Tags),
		})
	}

	feedName := f.Name
	if feedName == "" {
		feedName = respFeed.Title
	}

	feed := &Feed{
		ID:          uuid.UUID{},
		Config:      f,
		Title:       feedName,
		URL:         f.URL,
		Description: respFeed.Description,
		Items:       items,
		Key:         f.Key,
		Metadata:    make(map[string]interface{}),
		Tags:        f.Tags,
	}

	return feed, nil
}

func NewManager(repo repo) *Manager {
	return &Manager{
		repo: repo,
	}
}
