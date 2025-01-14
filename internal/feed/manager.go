package feed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"rssgram/internal/metrics"
	"rssgram/internal/storage"
	"rssgram/internal/utils"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"go.uber.org/zap"
)

type repo interface {
	GetFeedByURL(ctx context.Context, url string) (*storage.StoredFeed, error)
	DeleteFeed(ctx context.Context, url string) error
	UpsertFeed(ctx context.Context, url string, lastChecked, lastPost time.Time) error

	InsertItem(ctx context.Context, item *FeedItem) error
}

type Manager struct {
	repo repo
}

func (fm *Manager) EnrichFeedItems(feed *Feed) error {
	if feed.Config.DescriptionType != FeedDescriptionTypeLink {
		return nil
	}

	p := NewSiteParser()
	wg := sync.WaitGroup{}

	for i := range feed.Items {
		item := &feed.Items[i]

		wg.Add(1)
		go func() {
			defer wg.Done()
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

func (fm *Manager) ProcessFeed(ctx context.Context, f FeedConfig, logger *zap.Logger) error {
	ctxLogger := logger.With(zap.String("feed", f.Name), zap.String("url", f.URL))
	var isNewFeed bool

	storedFeed, err := fm.repo.GetFeedByURL(ctx, f.URL)
	if err != nil {
		return fmt.Errorf("failed get stored feed by url %s: %w", f.URL, err)
	}

	if storedFeed == nil {
		isNewFeed = true
		ctxLogger.Info("feed is new")
	}

	startTime := time.Now()
	feed, err := fm.GetFeed(ctx, f)
	metrics.FeedGetTimeSec.WithLabelValues(f.Name).Set(time.Since(startTime).Seconds())
	if err != nil {
		metrics.FeedGetError.WithLabelValues(f.Name).Inc()
		return fmt.Errorf("failed get feed by url %s: %w", f.URL, err)
	}

	metrics.FeedGetSuccess.WithLabelValues(f.Name).Inc()

	var lastItemPublishedAt time.Time
	var newItems int

	if isNewFeed {
		lastItemPublishedAt = fm.getMaxPublishedAt(feed)
	} else {
		feed.StoredLastSavedItem = storedFeed.LastPosted
		newItems, lastItemPublishedAt, err = fm.processFeed(ctx, feed, ctxLogger)
		if err != nil {
			return fmt.Errorf("failed process feed %s: %w", f.URL, err)
		}

		metrics.NewItemsCount.WithLabelValues(feed.Title).Add(float64(newItems))
	}

	err = fm.repo.UpsertFeed(ctx, f.URL, time.Now().UTC(), lastItemPublishedAt)
	if err != nil {
		return fmt.Errorf("failed upserting feed %s: %w", f.URL, err)
	}

	return nil
}

func (fm *Manager) processFeed(ctx context.Context, feed *Feed, ctxLogger *zap.Logger) (int, time.Time, error) {

	var newItemsAmount int

	// получаем время публикации последней новости
	lastItemPublishedAt := fm.getMaxPublishedAt(feed)
	ctxLogger.Debug(fmt.Sprintf("last published items: %v", lastItemPublishedAt))

	// отфильтровываем старые новости из фида
	newItemsAmount = fm.filterItemsAfterByRefTime(feed, feed.StoredLastSavedItem)
	ctxLogger.Debug(fmt.Sprintf("new items: %d", newItemsAmount))

	startTime := time.Now()
	if err := fm.EnrichFeedItems(feed); err != nil {
		metrics.ItemsEnrichTimeSec.WithLabelValues(feed.Title).Set(time.Since(startTime).Seconds())
		return newItemsAmount, lastItemPublishedAt, fmt.Errorf("failed enriching feed items (%s): %w", feed.URL, err)
	}

	for i := range feed.Items {
		err := fm.repo.InsertItem(context.Background(), &feed.Items[i])
		if err != nil {
			return newItemsAmount, lastItemPublishedAt, fmt.Errorf("failed inserting item %d: %w", i, err) // TODO: log instead break ???
		}

		ctxLogger.Debug(fmt.Sprintf("add new item: %s", feed.Items[i].ID))
	}

	return newItemsAmount, lastItemPublishedAt, nil
}

func (fm *Manager) getMaxPublishedAt(f *Feed) time.Time {
	var maxPublishedAt time.Time

	for i := range f.Items {
		if maxPublishedAt.IsZero() {
			maxPublishedAt = *f.Items[i].PublishedAt
		} else if maxPublishedAt.Before(*f.Items[i].PublishedAt) {
			maxPublishedAt = *f.Items[i].PublishedAt
		}
	}

	return maxPublishedAt
}

// TODO: add test
func (fm *Manager) filterItemsAfterByRefTime(f *Feed, refTime time.Time) int {
	var newItems []FeedItem

	for _, item := range f.Items {
		if item.PublishedAt.After(refTime) {
			newItems = append(newItems, item)
		}
	}

	f.Items = newItems

	return len(newItems)
}

func (fm *Manager) GetFeed(ctx context.Context, f FeedConfig) (*Feed, error) {

	fp := gofeed.NewParser()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	respFeed, err := fp.ParseURLWithContext(f.URL, ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to parse feed URL %s: %w", f.URL, err)
	}

	var items []FeedItem
	feedTitle := respFeed.Title
	if f.Name != "" {
		feedTitle = f.Name
	}

	for _, item := range respFeed.Items {

		var imageURL string
		if item.Image != nil {
			imageURL = item.Image.URL
		}

		items = append(items,
			NewFeedItem(
				feedTitle,
				item.Title,
				item.Link,
				imageURL,
				item.Description,
				item.PublishedParsed,
				item.UpdatedParsed,
				utils.MergeStrSlices(item.Categories, f.Tags),
			))
	}

	feed := &Feed{
		ID:          uuid.UUID{},
		Config:      f,
		Title:       feedTitle,
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
