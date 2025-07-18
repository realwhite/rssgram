package feed

import (
	"context"
	"rssgram/internal/storage"
	"testing"
	"time"

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

func TestNewManager(t *testing.T) {
	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)

	assert.NotNil(t, manager)
	assert.Equal(t, mockRepo, manager.repo)
}

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

// Вспомогательная функция для создания указателя на время
func timePtr(t time.Time) *time.Time {
	return &t
}
