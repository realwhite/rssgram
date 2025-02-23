package telegram

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

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
			o := NewTelegramChannelOutput(TelegramChannelOutputConfig{
				TelegramChannelClientConfig{
					ChannelName: "",
					BotToken:    "",
					SilentMode:  tc.silentConfig,
				},
			}, zap.NewNop())

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
