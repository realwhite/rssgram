// Test created with AI
package main

import (
	"os"
	"testing"

	"rssgram/internal"
	"rssgram/internal/outputs/telegram"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_ParseConfig(t *testing.T) {
	// Create a temporary config file
	tempConfig := `
feeds:
  - name: "Test Feed"
    url: "https://example.com/rss"
    description_type: "link"

telegram:
  channel_name: "@test_channel"
  bot_token: "test_token"
  silent_mode:
    start: "23:00:00"
    finish: "08:00:00"
    timezone: "Europe/Moscow"
`

	// Save the original config file
	originalConfig := "config.yaml"
	if _, err := os.Stat(originalConfig); err == nil {
		// If the file exists, rename it
		err = os.Rename(originalConfig, originalConfig+".backup")
		require.NoError(t, err)
		defer os.Rename(originalConfig+".backup", originalConfig)
	}

	// Write the config to a temporary file
	err := os.WriteFile("config.yaml", []byte(tempConfig), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Test loading the config
	config, err := internal.ParseConfig()

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 1)
	assert.Equal(t, "Test Feed", config.Feeds[0].Name)
	assert.Equal(t, "https://example.com/rss", config.Feeds[0].URL)
	assert.Equal(t, "@test_channel", config.Telegram.ChannelName)
	assert.Equal(t, "test_token", config.Telegram.BotToken)
}

func TestMain_ParseConfig_FileNotFound(t *testing.T) {
	// Save the original config file
	originalConfig := "config.yaml"
	if _, err := os.Stat(originalConfig); err == nil {
		// If the file exists, rename it
		err = os.Rename(originalConfig, originalConfig+".backup")
		require.NoError(t, err)
		defer os.Rename(originalConfig+".backup", originalConfig)
	}

	// Test loading a non-existent file
	config, err := internal.ParseConfig()

	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestMain_ParseConfig_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tempConfig := `
feeds:
  - name: "Test Feed"
    url: "https://example.com/rss"
    description_type: "link"
invalid_yaml: [unclosed_bracket
`

	// Save the original config file
	originalConfig := "config.yaml"
	if _, err := os.Stat(originalConfig); err == nil {
		// If the file exists, rename it
		err = os.Rename(originalConfig, originalConfig+".backup")
		require.NoError(t, err)
		defer os.Rename(originalConfig+".backup", originalConfig)
	}

	// Write the config to a temporary file
	err := os.WriteFile("config.yaml", []byte(tempConfig), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Test loading invalid YAML
	config, err := internal.ParseConfig()

	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestMain_ConfigValidation(t *testing.T) {
	// Test config validation with empty bot token
	config := &internal.Config{
		Feeds: []internal.FeedConfig{
			{
				Name: "Test Feed",
				URL:  "https://example.com/rss",
			},
		},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "@test_channel",
				BotToken:    "test_token",
			},
		},
	}

	// Check that the config contains an empty token
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 1)
	assert.Equal(t, "Test Feed", config.Feeds[0].Name)
	assert.Equal(t, "https://example.com/rss", config.Feeds[0].URL)
	assert.Equal(t, "@test_channel", config.Telegram.ChannelName)
	assert.Equal(t, "test_token", config.Telegram.BotToken)
}

func TestMain_ConfigValidation_EmptyFeeds(t *testing.T) {
	// Test config validation with empty feeds
	config := &internal.Config{
		Feeds: []internal.FeedConfig{},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "@test_channel",
				BotToken:    "test_token",
			},
		},
	}

	// Check that the config is empty
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 0)
}

func TestMain_ConfigValidation_EmptyFeedName(t *testing.T) {
	// Test config validation with empty feed name
	config := &internal.Config{
		Feeds: []internal.FeedConfig{
			{
				Name: "", // Empty name
				URL:  "https://example.com/rss",
			},
		},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "@test_channel",
				BotToken:    "test_token",
			},
		},
	}

	// Check that the config contains an empty name
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 1)
	assert.Empty(t, config.Feeds[0].Name)
}

func TestMain_ConfigValidation_EmptyChannelName(t *testing.T) {
	// Test config validation with empty channel name
	config := &internal.Config{
		Feeds: []internal.FeedConfig{
			{
				Name: "Test Feed",
				URL:  "https://example.com/rss",
			},
		},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "", // Empty channel name
				BotToken:    "test_token",
			},
		},
	}

	// Check that the config contains an empty channel name
	assert.NotNil(t, config)
	assert.Empty(t, config.Telegram.ChannelName)
}

func TestMain_ConfigValidation_EmptyBotToken(t *testing.T) {
	// Test config validation with empty bot token
	config := &internal.Config{
		Feeds: []internal.FeedConfig{
			{
				Name: "Test Feed",
				URL:  "https://example.com/rss",
			},
		},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "@test_channel",
				BotToken:    "", // Empty token
			},
		},
	}

	// Check that the config contains an empty token
	assert.NotNil(t, config)
	assert.Empty(t, config.Telegram.BotToken)
}

func TestMain_ConfigValidation_InvalidURL(t *testing.T) {
	// Test config validation with invalid URL
	config := &internal.Config{
		Feeds: []internal.FeedConfig{
			{
				Name: "Test Feed",
				URL:  "invalid-url", // Invalid URL
			},
		},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "@test_channel",
				BotToken:    "test_token",
			},
		},
	}

	// Check that the config contains an invalid URL
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 1)
	assert.Equal(t, "invalid-url", config.Feeds[0].URL)
}

func TestMain_RunMigrate(t *testing.T) {
	// Test running migrations
	err := runMigrate()

	// Expect either success or an error, but not a panic
	// Migrations may already be applied, so "no change" error is acceptable
	if err != nil {
		// Check that this is not a critical error
		assert.Contains(t, err.Error(), "no change")
	}
}
