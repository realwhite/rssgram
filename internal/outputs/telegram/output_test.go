package telegram

import (
	"testing"
	"time"

	"context"
	"rssgram/internal/feed"
)

// TestTelegramChannelOutput_IsSilentMode проверяет различные сценарии работы тихого режима (silent mode) для Telegram.
func TestTelegramChannelOutput_IsSilentMode(t *testing.T) {
	testCases := []struct {
		name         string
		silentConfig SilentModeConfig
		refTime      time.Time
		expected     bool
	}{
		{
			name: "empty",
			silentConfig: SilentModeConfig{
				Start:    "",
				Finish:   "",
				Timezone: "",
			},
			refTime:  time.Now(),
			expected: false,
		},
		{
			name: "only start",
			silentConfig: SilentModeConfig{
				Start:    "00:00:00",
				Finish:   "",
				Timezone: "",
			},
			refTime:  time.Now(),
			expected: false,
		},
		{
			name: "only finish",
			silentConfig: SilentModeConfig{
				Start:    "",
				Finish:   "00:00:00",
				Timezone: "",
			},
			refTime:  time.Now(),
			expected: false,
		},
		{
			name: "silent in one day (false)",
			silentConfig: SilentModeConfig{
				Start:    "10:00:00",
				Finish:   "15:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 9, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name: "silent in one day (true)",
			silentConfig: SilentModeConfig{
				Start:    "10:00:00",
				Finish:   "15:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 10, 0, 1, 0, time.UTC),
			expected: true,
		},
		{
			name: "silent in one day (true)",
			silentConfig: SilentModeConfig{
				Start:    "10:00:00",
				Finish:   "15:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 12, 12, 12, 0, time.UTC),
			expected: true,
		},
		{
			name: "silent in one day (false)",
			silentConfig: SilentModeConfig{
				Start:    "10:00:00",
				Finish:   "15:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 15, 00, 01, 0, time.UTC),
			expected: false,
		},
		{
			name: "silent in one day (false)",
			silentConfig: SilentModeConfig{
				Start:    "10:00:00",
				Finish:   "15:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 15, 1, 00, 0, time.UTC),
			expected: false,
		},
		{
			name: "silent in two days (false)",
			silentConfig: SilentModeConfig{
				Start:    "23:00:00",
				Finish:   "06:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 22, 59, 00, 0, time.UTC),
			expected: false,
		},
		{
			name: "silent in two days (true)",
			silentConfig: SilentModeConfig{
				Start:    "23:00:00",
				Finish:   "06:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 23, 0, 1, 0, time.UTC),
			expected: true,
		},
		{
			name: "silent in two days (true)",
			silentConfig: SilentModeConfig{
				Start:    "23:00:00",
				Finish:   "06:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 00, 00, 00, 0, time.UTC),
			expected: true,
		},
		{
			name: "silent in two days (true)",
			silentConfig: SilentModeConfig{
				Start:    "23:00:00",
				Finish:   "06:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 1, 00, 00, 0, time.UTC),
			expected: true,
		},
		{
			name: "silent in two days (true)",
			silentConfig: SilentModeConfig{
				Start:    "23:00:00",
				Finish:   "06:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 5, 59, 59, 0, time.UTC),
			expected: true,
		},
		{
			name: "silent in two days (true)",
			silentConfig: SilentModeConfig{
				Start:    "23:00:00",
				Finish:   "06:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 6, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name: "silent in two days (false)",
			silentConfig: SilentModeConfig{
				Start:    "23:00:00",
				Finish:   "06:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 6, 0, 1, 0, time.UTC),
			expected: false,
		},
		{
			name: "edge1",
			silentConfig: SilentModeConfig{
				Start:    "23:00:00",
				Finish:   "06:00:00",
				Timezone: "",
			},
			refTime:  time.Date(2025, 10, 17, 0, 50, 1, 0, time.UTC),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o := &TelegramChannelOutput{
				config: TelegramChannelOutputConfig{
					TelegramChannelClientConfig: TelegramChannelClientConfig{
						SilentMode: tc.silentConfig,
					},
				},
			}
			res, err := o.IsSilentMode(tc.silentConfig.Start, tc.silentConfig.Finish, tc.silentConfig.Timezone, tc.refTime)
			if err != nil {
				t.Errorf("IsSilentMode() error = %v", err)
			}

			if res != tc.expected {
				t.Errorf("IsSilentMode() = %v, want %v", res, tc.expected)
			}
		})
	}
}

// Вспомогательный мок-клиент

type mockTelegramChannelClient struct {
	sendMessageFunc func(ctx context.Context, msg string, options TelegramMessageOptions) error
	sendPhotoFunc   func(ctx context.Context, msg, photoUrl string, disableNotification bool) error
}

func (m *mockTelegramChannelClient) SendMessage(ctx context.Context, msg string, options TelegramMessageOptions) error {
	return m.sendMessageFunc(ctx, msg, options)
}
func (m *mockTelegramChannelClient) SendPhoto(ctx context.Context, msg, photoUrl string, disableNotification bool) error {
	return m.sendPhotoFunc(ctx, msg, photoUrl, disableNotification)
}

// TestTelegramChannelOutput_Push_Tags проверяет, что теги корректно добавляются или не добавляются в сообщение в зависимости от флага enableTags и наличия тегов.
func TestTelegramChannelOutput_Push_Tags(t *testing.T) {
	testCases := []struct {
		enableTags bool
		tags       []string
		expectIn   string
		name       string
	}{
		{true, []string{"go", "news"}, "#go #news", "Теги включены, несколько тегов"},
		{true, []string{"один"}, "#один", "Теги включены, один тег"},
		{false, []string{"go", "news"}, "#go #news", "Теги выключены, не должно быть тегов"},
		{true, nil, "", "Теги включены, но тегов нет"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &TelegramChannelOutput{
				client:     nil, // будет заменён
				config:     TelegramChannelOutputConfig{},
				enableTags: tc.enableTags,
			}

			item := &feed.FeedItem{
				FeedTitle:   "TestFeed",
				Title:       "TestTitle",
				Link:        "https://example.com",
				Description: "desc",
				Tags:        tc.tags,
			}

			var sentMsg string
			output.client = &mockTelegramChannelClient{
				sendMessageFunc: func(ctx context.Context, msg string, options TelegramMessageOptions) error {
					sentMsg = msg
					return nil
				},
				sendPhotoFunc: func(ctx context.Context, msg, photoUrl string, disableNotification bool) error {
					sentMsg = msg
					return nil
				},
			}

			item.ImageURL = "" // тестируем SendMessage
			_, _ = output.Push(context.Background(), item)

			if tc.enableTags && len(tc.tags) > 0 {
				if !contains(sentMsg, tc.expectIn) {
					t.Errorf("Ожидал теги в сообщении: %q, got: %q", tc.expectIn, sentMsg)
				}
			} else {
				if contains(sentMsg, "#") {
					t.Errorf("Не ожидал тегов в сообщении, got: %q", sentMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return substr == "" || (len(substr) > 0 && (len(s) >= len(substr)) && (s == substr || (len(s) > len(substr) && (s[len(s)-len(substr):] == substr || s[len(s)-len(substr)-1:] == "\n"+substr)))) || (len(substr) > 0 && (len(s) > len(substr)) && (s[len(s)-len(substr)-2:] == "\n\n"+substr)) || (len(substr) > 0 && (len(s) > len(substr)) && (s[len(s)-len(substr)-1:] == " "+substr))
}
