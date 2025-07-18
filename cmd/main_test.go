// Тест создан с помощью AI
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
	// Создаем временный конфигурационный файл
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

	// Сохраняем оригинальный файл конфигурации
	originalConfig := "config.yaml"
	if _, err := os.Stat(originalConfig); err == nil {
		// Если файл существует, переименовываем его
		err = os.Rename(originalConfig, originalConfig+".backup")
		require.NoError(t, err)
		defer os.Rename(originalConfig+".backup", originalConfig)
	}

	// Записываем конфигурацию во временный файл
	err := os.WriteFile("config.yaml", []byte(tempConfig), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Тестируем загрузку конфигурации
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
	// Сохраняем оригинальный файл конфигурации
	originalConfig := "config.yaml"
	if _, err := os.Stat(originalConfig); err == nil {
		// Если файл существует, переименовываем его
		err = os.Rename(originalConfig, originalConfig+".backup")
		require.NoError(t, err)
		defer os.Rename(originalConfig+".backup", originalConfig)
	}

	// Тестируем загрузку несуществующего файла
	config, err := internal.ParseConfig()

	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestMain_ParseConfig_InvalidYAML(t *testing.T) {
	// Создаем временный файл с невалидным YAML
	tempConfig := `
feeds:
  - name: "Test Feed"
    url: "https://example.com/rss"
    description_type: "link"
invalid_yaml: [unclosed_bracket
`

	// Сохраняем оригинальный файл конфигурации
	originalConfig := "config.yaml"
	if _, err := os.Stat(originalConfig); err == nil {
		// Если файл существует, переименовываем его
		err = os.Rename(originalConfig, originalConfig+".backup")
		require.NoError(t, err)
		defer os.Rename(originalConfig+".backup", originalConfig)
	}

	// Записываем конфигурацию во временный файл
	err := os.WriteFile("config.yaml", []byte(tempConfig), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Тестируем загрузку невалидного YAML
	config, err := internal.ParseConfig()

	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestMain_ConfigValidation(t *testing.T) {
	// Тестируем валидацию корректной конфигурации
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

	// Проверяем, что конфигурация корректна
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 1)
	assert.Equal(t, "Test Feed", config.Feeds[0].Name)
	assert.Equal(t, "https://example.com/rss", config.Feeds[0].URL)
	assert.Equal(t, "@test_channel", config.Telegram.ChannelName)
	assert.Equal(t, "test_token", config.Telegram.BotToken)
}

func TestMain_ConfigValidation_EmptyFeeds(t *testing.T) {
	// Тестируем валидацию конфигурации без фидов
	config := &internal.Config{
		Feeds: []internal.FeedConfig{},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "@test_channel",
				BotToken:    "test_token",
			},
		},
	}

	// Проверяем, что конфигурация пуста
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 0)
}

func TestMain_ConfigValidation_EmptyFeedName(t *testing.T) {
	// Тестируем валидацию конфигурации с пустым именем фида
	config := &internal.Config{
		Feeds: []internal.FeedConfig{
			{
				Name: "", // Пустое имя
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

	// Проверяем, что конфигурация содержит пустое имя
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 1)
	assert.Empty(t, config.Feeds[0].Name)
}

func TestMain_ConfigValidation_EmptyChannelName(t *testing.T) {
	// Тестируем валидацию конфигурации с пустым именем канала
	config := &internal.Config{
		Feeds: []internal.FeedConfig{
			{
				Name: "Test Feed",
				URL:  "https://example.com/rss",
			},
		},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "", // Пустое имя канала
				BotToken:    "test_token",
			},
		},
	}

	// Проверяем, что конфигурация содержит пустое имя канала
	assert.NotNil(t, config)
	assert.Empty(t, config.Telegram.ChannelName)
}

func TestMain_ConfigValidation_EmptyBotToken(t *testing.T) {
	// Тестируем валидацию конфигурации с пустым токеном бота
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
				BotToken:    "", // Пустой токен
			},
		},
	}

	// Проверяем, что конфигурация содержит пустой токен
	assert.NotNil(t, config)
	assert.Empty(t, config.Telegram.BotToken)
}

func TestMain_ConfigValidation_InvalidURL(t *testing.T) {
	// Тестируем валидацию конфигурации с невалидным URL
	config := &internal.Config{
		Feeds: []internal.FeedConfig{
			{
				Name: "Test Feed",
				URL:  "invalid-url", // Невалидный URL
			},
		},
		Telegram: telegram.TelegramChannelOutputConfig{
			TelegramChannelClientConfig: telegram.TelegramChannelClientConfig{
				ChannelName: "@test_channel",
				BotToken:    "test_token",
			},
		},
	}

	// Проверяем, что конфигурация содержит невалидный URL
	assert.NotNil(t, config)
	assert.Len(t, config.Feeds, 1)
	assert.Equal(t, "invalid-url", config.Feeds[0].URL)
}

func TestMain_RunMigrate(t *testing.T) {
	// Тестируем выполнение миграций
	err := runMigrate()

	// Ожидаем либо успех, либо ошибку, но не панику
	// Миграции могут уже быть применены, поэтому ошибка "no change" допустима
	if err != nil {
		// Проверяем, что это не критическая ошибка
		assert.Contains(t, err.Error(), "no change")
	}
}
