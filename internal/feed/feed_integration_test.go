// Test created with AI
//go:build integration
// +build integration

package feed

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestFeedManager_GetFeed_RealRSS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)
	logger := zap.NewNop()

	tests := []struct {
		name        string
		feedConfig  FeedConfig
		expectError bool
	}{
		{
			name: "Hacker News RSS",
			feedConfig: FeedConfig{
				Name: "Hacker News",
				URL:  "https://news.ycombinator.com/rss",
			},
			expectError: false,
		},
		{
			name: "Invalid RSS URL",
			feedConfig: FeedConfig{
				Name: "Invalid Feed",
				URL:  "https://example.com/nonexistent-rss",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			feed, err := manager.GetFeed(ctx, tt.feedConfig)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, feed)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, feed)
				assert.NotEmpty(t, feed.Title)
				assert.NotEmpty(t, feed.URL)
				assert.Greater(t, len(feed.Items), 0)

				// Check the structure of the first item
				if len(feed.Items) > 0 {
					item := feed.Items[0]
					assert.NotEmpty(t, item.ID)
					assert.NotEmpty(t, item.Title)
					assert.NotEmpty(t, item.Link)
					assert.Equal(t, feed.Title, item.FeedTitle)
				}
			}
		})
	}
}

func TestFeedManager_ProcessFeed_RealRSS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)
	logger := zap.NewNop()

	feedConfig := FeedConfig{
		Name: "Test RSS",
		URL:  "https://news.ycombinator.com/rss",
	}

	// Mock GetFeedByURL for the new feed
	mockRepo.On("GetFeedByURL", mock.Anything, feedConfig.URL).Return(nil, nil)
	mockRepo.On("UpsertFeed", mock.Anything, feedConfig.URL, mock.Anything, mock.Anything).Return(nil)

	// Test ProcessFeed with real RSS
	err := manager.ProcessFeed(context.Background(), feedConfig, logger)

	// Expect success for real RSS
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFeedManager_EnrichFeedItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mockRepo := &MockRepo{}
	manager := NewManager(mockRepo)

	// Create a test feed with a description type of "link"
	feed := &Feed{
		Config: FeedConfig{
			DescriptionType: FeedDescriptionTypeLink,
		},
		Items: []FeedItem{
			{
				Title: "Test Article",
				Link:  "https://example.com/article",
			},
		},
	}

	// Test enriching items
	err := manager.EnrichFeedItems(feed)
	assert.NoError(t, err)
}
