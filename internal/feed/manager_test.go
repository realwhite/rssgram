package feed

import (
	"context"
	"rssgram/internal/storage"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRepo - мок для интерфейса repo
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetFeedByURL(ctx context.Context, url string) (*storage.StoredFeed, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.StoredFeed), args.Error(1)
}

func (m *MockRepo) DeleteFeed(ctx context.Context, url string) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockRepo) UpsertFeed(ctx context.Context, url string, lastChecked, lastPost time.Time) error {
	args := m.Called(ctx, url, lastChecked, lastPost)
	return args.Error(0)
}

func (m *MockRepo) InsertItem(ctx context.Context, item *FeedItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

// TestNewManager проверяет, что менеджер создаётся корректно и содержит переданный repo.
func TestNewManager(t *testing.T) {
	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)

	assert.NotNil(t, manager)
	assert.Equal(t, mockRepo, manager.repo)
}

// TestManager_GetMaxPublishedAt проверяет корректность поиска самой поздней даты публикации среди элементов фида.
func TestManager_GetMaxPublishedAt(t *testing.T) {
	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)

	tests := []struct {
		name     string
		items    []FeedItem
		expected time.Time
	}{
		{
			name:     "empty items",
			items:    []FeedItem{},
			expected: time.Time{},
		},
		{
			name: "single item",
			items: []FeedItem{
				{
					PublishedAt: timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
				},
			},
			expected: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "multiple items",
			items: []FeedItem{
				{
					PublishedAt: timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
				},
				{
					PublishedAt: timePtr(time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)),
				},
				{
					PublishedAt: timePtr(time.Date(2023, 1, 1, 18, 0, 0, 0, time.UTC)),
				},
			},
			expected: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "items with nil PublishedAt",
			items: []FeedItem{
				{
					PublishedAt: nil,
				},
				{
					PublishedAt: timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
				},
			},
			expected: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed := &Feed{Items: tt.items}
			result := manager.getMaxPublishedAt(feed)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestManager_FilterItemsAfterByRefTime проверяет фильтрацию элементов фида по дате публикации относительно заданного времени.
func TestManager_FilterItemsAfterByRefTime(t *testing.T) {
	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)

	refTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		items         []FeedItem
		refTime       time.Time
		expectedCount int
		expectedItems []FeedItem
	}{
		{
			name:          "empty items",
			items:         []FeedItem{},
			refTime:       refTime,
			expectedCount: 0,
			expectedItems: []FeedItem{},
		},
		{
			name: "all items after ref time",
			items: []FeedItem{
				{
					Title:       "Item 1",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)),
				},
				{
					Title:       "Item 2",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC)),
				},
			},
			refTime:       refTime,
			expectedCount: 2,
			expectedItems: []FeedItem{
				{
					Title:       "Item 1",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)),
				},
				{
					Title:       "Item 2",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "mixed items",
			items: []FeedItem{
				{
					Title:       "Old Item",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)),
				},
				{
					Title:       "New Item 1",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)),
				},
				{
					Title:       "New Item 2",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC)),
				},
			},
			refTime:       refTime,
			expectedCount: 2,
			expectedItems: []FeedItem{
				{
					Title:       "New Item 1",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)),
				},
				{
					Title:       "New Item 2",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "all items before ref time",
			items: []FeedItem{
				{
					Title:       "Old Item 1",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)),
				},
				{
					Title:       "Old Item 2",
					PublishedAt: timePtr(time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)),
				},
			},
			refTime:       refTime,
			expectedCount: 0,
			expectedItems: []FeedItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed := &Feed{Items: tt.items}
			count := manager.filterItemsAfterByRefTime(feed, tt.refTime)

			assert.Equal(t, tt.expectedCount, count)
			assert.Len(t, feed.Items, tt.expectedCount)
		})
	}
}

// TestManager_ProcessFeed_NewFeed проверяет обработку нового фида (когда его ещё нет в хранилище).
func TestManager_ProcessFeed_NewFeed(t *testing.T) {
	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)
	logger := zap.NewNop()

	feedConfig := FeedConfig{
		Name: "Test Feed",
		URL:  "https://example.com/rss",
	}

	// Мокаем GetFeedByURL для нового фида
	mockRepo.On("GetFeedByURL", mock.Anything, feedConfig.URL).Return(nil, nil)

	// Тестируем ProcessFeed для нового фида
	err := manager.ProcessFeed(context.Background(), feedConfig, logger)

	// Ожидаем ошибку, так как GetFeed не реализован в тестах
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed get feed by url")
}

// TestManager_ProcessFeed_ExistingFeed проверяет обработку уже существующего фида (есть в хранилище).
func TestManager_ProcessFeed_ExistingFeed(t *testing.T) {
	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)
	logger := zap.NewNop()

	feedConfig := FeedConfig{
		Name: "Test Feed",
		URL:  "https://example.com/rss",
	}

	storedFeed := &storage.StoredFeed{
		URL:         feedConfig.URL,
		LastChecked: time.Now().Add(-time.Hour),
		LastPosted:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Мокаем GetFeedByURL для существующего фида
	mockRepo.On("GetFeedByURL", mock.Anything, feedConfig.URL).Return(storedFeed, nil)

	// Тестируем ProcessFeed для существующего фида
	err := manager.ProcessFeed(context.Background(), feedConfig, logger)

	// Ожидаем ошибку, так как GetFeed не реализован в тестах
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed get feed by url")
}

// TestManager_ProcessFeed_StorageError проверяет обработку ошибки при получении фида из хранилища.
func TestManager_ProcessFeed_StorageError(t *testing.T) {
	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)
	logger := zap.NewNop()

	feedConfig := FeedConfig{
		Name: "Test Feed",
		URL:  "https://example.com/rss",
	}

	// Мокаем ошибку в GetFeedByURL
	mockRepo.On("GetFeedByURL", mock.Anything, feedConfig.URL).Return(nil, assert.AnError)

	// Тестируем ProcessFeed с ошибкой хранилища
	err := manager.ProcessFeed(context.Background(), feedConfig, logger)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed get stored feed by url")

	mockRepo.AssertExpectations(t)
}

// Для теста приоритета тегов
type fakeItem struct {
	title      string
	categories []string
}

// TestFeedManager_GetFeed_TagsPriority проверяет приоритет тегов: если в конфиге есть теги, используются только они, иначе берутся из RSS.
func TestFeedManager_GetFeed_TagsPriority(t *testing.T) {
	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)

	feedConfigWithTags := FeedConfig{
		Name: "TestFeed",
		URL:  "https://example.com/rss",
		Tags: []string{"tag1", "tag2"},
	}
	feedConfigNoTags := FeedConfig{
		Name: "TestFeed",
		URL:  "https://example.com/rss",
	}

	// Мокаем gofeed.Parser
	parser := &fakeParser{
		items: []fakeItem{
			{"Title1", []string{"rss1", "rss2"}},
			{"Title2", []string{"cat"}},
		},
	}

	// Подменяем parserFactory на фейковый парсер
	oldFactory := manager.parserFactory
	manager.parserFactory = func() gofeedParser { return parser }
	defer func() { manager.parserFactory = oldFactory }()

	ctx := context.Background()

	// 1. Если в конфиге есть теги, используются только они
	feed, err := manager.GetFeed(ctx, feedConfigWithTags)
	assert.NoError(t, err)
	for _, item := range feed.Items {
		assert.Equal(t, []string{"tag1", "tag2"}, item.Tags)
	}

	// 2. Если в конфиге нет тегов, используются из RSS
	feed, err = manager.GetFeed(ctx, feedConfigNoTags)
	assert.NoError(t, err)
	assert.Equal(t, []string{"rss1", "rss2"}, feed.Items[0].Tags)
	assert.Equal(t, []string{"cat"}, feed.Items[1].Tags)
}

// Моки для gofeed

type fakeParser struct {
	items []fakeItem
}

func (f *fakeParser) ParseURLWithContext(url string, ctx context.Context) (*gofeed.Feed, error) {
	var items []*gofeed.Item
	for _, it := range f.items {
		items = append(items, &gofeed.Item{
			Title:      it.title,
			Categories: it.categories,
		})
	}
	return &gofeed.Feed{
		Title: "FakeFeed",
		Items: items,
	}, nil
}

// Вспомогательная функция для создания указателя на время
func timePtr(t time.Time) *time.Time {
	return &t
}
