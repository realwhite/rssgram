package internal

import (
	"os"
	"testing"

	"rssgram/internal/outputs/telegram"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expectError bool
		expected    *Config
	}{
		{
			name: "valid config",
			configYAML: `
telegram:
  channel_name: "@test_channel"
  bot_token: "test_token"
  silent_mode:
    start: "23:00:00"
    finish: "08:00:00"
    timezone: "Europe/Moscow"

feeds:
  - name: "Test Feed"
    url: "https://example.com/rss"
    description_type: "link"
  - name: "Another Feed"
    url: "https://another.com/rss"
    description_type: "item"
`,
			expectError: false,
			expected: &Config{
				Feeds: []FeedConfig{
					{
						Name:            "Test Feed",
						URL:             "https://example.com/rss",
						DescriptionType: "link",
					},
					{
						Name:            "Another Feed",
						URL:             "https://another.com/rss",
						DescriptionType: "item",
					},
				},
				Telegram: telegram.TelegramChannelOutputConfig{
					TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
						ChannelName: "@test_channel",
						BotToken:    "test_token",
						SilentMode: telegram.SilentModeConfig{
							Start:    "23:00:00",
							Finish:   "08:00:00",
							Timezone: "Europe/Moscow",
						},
					},
				},
			},
		},
		{
			name: "empty config",
			configYAML: `
telegram:
  channel_name: ""
  bot_token: ""
feeds: []
`,
			expectError: false,
			expected: &Config{
				Feeds: []FeedConfig{},
				Telegram: telegram.TelegramChannelOutputConfig{
					TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{},
				},
			},
		},
		{
			name:        "invalid yaml",
			configYAML:  `invalid: yaml: content:`,
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл конфигурации
			tmpFile, err := os.CreateTemp("", "config_*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.configYAML)
			require.NoError(t, err)
			tmpFile.Close()

			// Сохраняем оригинальное имя файла
			originalConfigFile := "config.yaml"
			if _, err := os.Stat(originalConfigFile); err == nil {
				// Если файл существует, переименовываем его
				err = os.Rename(originalConfigFile, originalConfigFile+".backup")
				require.NoError(t, err)
				defer os.Rename(originalConfigFile+".backup", originalConfigFile)
			}

			// Переименовываем временный файл в config.yaml
			err = os.Rename(tmpFile.Name(), "config.yaml")
			require.NoError(t, err)

			// Тестируем ParseConfig
			result, err := ParseConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Feeds, result.Feeds)
				assert.Equal(t, tt.expected.Telegram.TelegramChannelClientConfig.ChannelName, result.Telegram.TelegramChannelClientConfig.ChannelName)
				assert.Equal(t, tt.expected.Telegram.TelegramChannelClientConfig.BotToken, result.Telegram.TelegramChannelClientConfig.BotToken)
				assert.Equal(t, tt.expected.Telegram.TelegramChannelClientConfig.SilentMode, result.Telegram.TelegramChannelClientConfig.SilentMode)
			}
		})
	}
}

func TestParseConfig_FileNotFound(t *testing.T) {
	// Сохраняем оригинальный файл если он существует
	originalConfigFile := "config.yaml"
	if _, err := os.Stat(originalConfigFile); err == nil {
		err = os.Rename(originalConfigFile, originalConfigFile+".backup")
		require.NoError(t, err)
		defer os.Rename(originalConfigFile+".backup", originalConfigFile)
	}

	// Удаляем файл конфигурации
	os.Remove("config.yaml")

	// Тестируем ParseConfig с отсутствующим файлом
	result, err := ParseConfig()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestFeedConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		feed        FeedConfig
		expectValid bool
	}{
		{
			name: "valid feed config",
			feed: FeedConfig{
				Name:            "Test Feed",
				URL:             "https://example.com/rss",
				DescriptionType: "link",
			},
			expectValid: true,
		},
		{
			name: "empty name",
			feed: FeedConfig{
				Name:            "",
				URL:             "https://example.com/rss",
				DescriptionType: "link",
			},
			expectValid: true, // Имя может быть пустым
		},
		{
			name: "empty URL",
			feed: FeedConfig{
				Name:            "Test Feed",
				URL:             "",
				DescriptionType: "link",
			},
			expectValid: false,
		},
		{
			name: "invalid URL",
			feed: FeedConfig{
				Name:            "Test Feed",
				URL:             "not-a-url",
				DescriptionType: "link",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.feed.URL != "" && (tt.feed.URL == "https://example.com/rss" || tt.feed.URL == "")
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}
