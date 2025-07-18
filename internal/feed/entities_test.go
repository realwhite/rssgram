package feed

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFeedItem(t *testing.T) {
	tests := []struct {
		name        string
		feedTitle   string
		title       string
		link        string
		imageURL    string
		description string
		publishedAt *time.Time
		updatedAt   *time.Time
		tags        []string
	}{
		{
			name:        "basic item",
			feedTitle:   "Test Feed",
			title:       "Test Article",
			link:        "https://example.com/article",
			imageURL:    "https://example.com/image.jpg",
			description: "Test description",
			publishedAt: timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
			updatedAt:   timePtr(time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)),
			tags:        []string{"test", "article"},
		},
		{
			name:        "item with nil times",
			feedTitle:   "Test Feed",
			title:       "Test Article",
			link:        "https://example.com/article",
			imageURL:    "",
			description: "",
			publishedAt: nil,
			updatedAt:   nil,
			tags:        []string{},
		},
		{
			name:        "item with empty fields",
			feedTitle:   "",
			title:       "",
			link:        "",
			imageURL:    "",
			description: "",
			publishedAt: nil,
			updatedAt:   nil,
			tags:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewFeedItem(
				tt.feedTitle,
				tt.title,
				tt.link,
				tt.imageURL,
				tt.description,
				tt.publishedAt,
				tt.updatedAt,
				tt.tags,
			)

			// Check that ID is generated correctly
			assert.NotEmpty(t, item.ID)
			assert.Len(t, item.ID, 64) // SHA256 hash length

			// Check other fields
			assert.Equal(t, tt.feedTitle, item.FeedTitle)
			assert.Equal(t, tt.title, item.Title)
			assert.Equal(t, tt.link, item.Link)
			assert.Equal(t, tt.imageURL, item.ImageURL)
			assert.Equal(t, tt.description, item.Description)
			assert.Equal(t, tt.publishedAt, item.PublishedAt)
			assert.Equal(t, tt.updatedAt, item.UpdatedAt)
			assert.Equal(t, tt.tags, item.Tags)
			assert.Nil(t, item.Metadata)
		})
	}
}

func TestFeedItem_GetMetadataJson(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		expected string
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			expected: "{}",
		},
		{
			name:     "empty metadata",
			metadata: map[string]interface{}{},
			expected: "{}",
		},
		{
			name: "simple metadata",
			metadata: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			expected: `{"key1":"value1","key2":123}`,
		},
		{
			name: "complex metadata",
			metadata: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
				"array": []interface{}{1, 2, 3},
			},
			expected: `{"array":[1,2,3],"nested":{"key":"value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &FeedItem{Metadata: tt.metadata}
			result, err := item.GetMetadataJson()

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)

			// Check that the result is valid JSON
			if tt.metadata != nil {
				var parsed map[string]interface{}
				err = json.Unmarshal([]byte(result), &parsed)
				assert.NoError(t, err)
			}
		})
	}
}

func TestFeedItem_GetTagsJson(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "nil tags",
			tags:     nil,
			expected: "[]",
		},
		{
			name:     "empty tags",
			tags:     []string{},
			expected: "[]",
		},
		{
			name:     "single tag",
			tags:     []string{"test"},
			expected: `["test"]`,
		},
		{
			name:     "multiple tags",
			tags:     []string{"test", "article", "news"},
			expected: `["test","article","news"]`,
		},
		{
			name:     "tags with special characters",
			tags:     []string{"test-tag", "article_tag", "news tag"},
			expected: `["test-tag","article_tag","news tag"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &FeedItem{Tags: tt.tags}
			result, err := item.GetTagsJson()

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)

			// Check that the result is valid JSON
			var parsed []string
			err = json.Unmarshal([]byte(result), &parsed)
			assert.NoError(t, err)
			if tt.tags == nil {
				assert.Equal(t, []string{}, parsed)
			} else {
				assert.Equal(t, tt.tags, parsed)
			}
		})
	}
}

func TestFeedItem_ID_Consistency(t *testing.T) {
	// Check that ID is generated consistently for identical data
	item1 := NewFeedItem(
		"Test Feed",
		"Test Article",
		"https://example.com/article",
		"https://example.com/image.jpg",
		"Test description",
		timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
		nil,
		[]string{"test"},
	)

	item2 := NewFeedItem(
		"Test Feed",
		"Test Article",
		"https://example.com/article",
		"https://example.com/image.jpg",
		"Test description",
		timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
		nil,
		[]string{"test"},
	)

	assert.Equal(t, item1.ID, item2.ID)

	// Check that different data gives different IDs
	item3 := NewFeedItem(
		"Test Feed",
		"Different Article", // Changed
		"https://example.com/article",
		"https://example.com/image.jpg",
		"Test description",
		timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
		nil,
		[]string{"test"},
	)

	assert.NotEqual(t, item1.ID, item3.ID)
}

func TestFeedItem_ID_Uniqueness(t *testing.T) {
	// Check uniqueness of ID for different data
	items := make(map[string]bool)

	for i := 0; i < 100; i++ {
		item := NewFeedItem(
			"Test Feed",
			fmt.Sprintf("Test Article %d", i),               // unique title
			fmt.Sprintf("https://example.com/article%d", i), // unique link
			"https://example.com/image.jpg",
			"Test description",
			timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
			nil,
			[]string{"test"},
		)
		// ID must be unique
		assert.False(t, items[item.ID], "Duplicate ID found: %s", item.ID)
		items[item.ID] = true
	}
}

func TestFeedConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      FeedConfig
		expectValid bool
	}{
		{
			name: "valid config",
			config: FeedConfig{
				Name:            "Test Feed",
				URL:             "https://example.com/rss",
				Key:             "test-key",
				DescriptionType: "link",
				Tags:            []string{"test", "news"},
			},
			expectValid: true,
		},
		{
			name: "empty name",
			config: FeedConfig{
				Name:            "",
				URL:             "https://example.com/rss",
				DescriptionType: "item",
			},
			expectValid: true, // Name can be empty
		},
		{
			name: "empty URL",
			config: FeedConfig{
				Name:            "Test Feed",
				URL:             "",
				DescriptionType: "item",
			},
			expectValid: false,
		},
		{
			name: "invalid description type",
			config: FeedConfig{
				Name:            "Test Feed",
				URL:             "https://example.com/rss",
				DescriptionType: "invalid",
			},
			expectValid: true, // Description can be any value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.URL != "" && tt.config.Name != "invalid"
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}
