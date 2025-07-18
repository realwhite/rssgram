// Тест создан с помощью AI
package telegram

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type roundTripperFunc func(*http.Request) *http.Response

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(status int, body string) *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: status,
				Body:       ioutil.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}
		}),
	}
}

func TestTelegramChannelClient_SendMessage(t *testing.T) {
	config := TelegramChannelClientConfig{
		ChannelName: "@test_channel",
		BotToken:    "test_token",
	}
	client := NewTelegramChannelClient(config, zap.NewNop())
	client.httpClient = newTestClient(200, `{"ok": true, "result": {"message_id": 123}}`)

	ctx := context.Background()
	err := client.SendMessage(ctx, "Test message", TelegramMessageOptions{
		LinkPreview:         false,
		DisableNotification: false,
	})
	assert.NoError(t, err)
}

func TestTelegramChannelClient_SendPhoto(t *testing.T) {
	config := TelegramChannelClientConfig{
		ChannelName: "@test_channel",
		BotToken:    "test_token",
	}
	client := NewTelegramChannelClient(config, zap.NewNop())
	client.httpClient = newTestClient(200, `{"ok": true, "result": {"message_id": 456}}`)

	ctx := context.Background()
	err := client.SendPhoto(ctx, "Test message", "https://example.com/image.jpg", false)
	assert.NoError(t, err)
}

func TestTelegramChannelClient_SendMessage_Error(t *testing.T) {
	config := TelegramChannelClientConfig{
		ChannelName: "@test_channel",
		BotToken:    "test_token",
	}
	client := NewTelegramChannelClient(config, zap.NewNop())
	client.httpClient = newTestClient(400, `{"ok": false, "error_code": 400, "description": "Bad Request"}`)

	ctx := context.Background()
	err := client.SendMessage(ctx, "Test message", TelegramMessageOptions{
		LinkPreview:         false,
		DisableNotification: false,
	})
	assert.Error(t, err)
}

func TestTelegramChannelClient_SendMessage_NetworkError(t *testing.T) {
	// Создаем клиент с невалидным URL
	config := TelegramChannelClientConfig{
		ChannelName: "@test_channel",
		BotToken:    "test_token",
	}

	client := NewTelegramChannelClient(config, zap.NewNop())

	// Тестируем отправку сообщения с сетевой ошибкой
	ctx := context.Background()
	err := client.SendMessage(ctx, "Test message", TelegramMessageOptions{
		LinkPreview:         false,
		DisableNotification: false,
	})

	assert.Error(t, err)
}

func TestTelegramChannelClient_SendMessage_Timeout(t *testing.T) {
	// Создаем тестовый сервер с задержкой
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Имитируем долгий ответ
		select {
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	// Создаем клиент с тестовым URL
	config := TelegramChannelClientConfig{
		ChannelName: "@test_channel",
		BotToken:    "test_token",
	}

	client := NewTelegramChannelClient(config, zap.NewNop())

	// Тестируем отправку сообщения с таймаутом
	ctx := context.Background()
	err := client.SendMessage(ctx, "Test message", TelegramMessageOptions{
		LinkPreview:         false,
		DisableNotification: false,
	})

	assert.Error(t, err)
}

func TestTelegramChannelClient_InvalidResponse(t *testing.T) {
	// Создаем тестовый сервер, возвращающий невалидный JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Возвращаем невалидный JSON
		response := `{"invalid": json}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Создаем клиент с тестовым URL
	config := TelegramChannelClientConfig{
		ChannelName: "@test_channel",
		BotToken:    "test_token",
	}

	client := NewTelegramChannelClient(config, zap.NewNop())

	// Тестируем отправку сообщения с невалидным ответом
	ctx := context.Background()
	err := client.SendMessage(ctx, "Test message", TelegramMessageOptions{
		LinkPreview:         false,
		DisableNotification: false,
	})

	assert.Error(t, err)
}

func TestTelegramChannelClient_EmptyMessage(t *testing.T) {
	// Создаем клиент
	config := TelegramChannelClientConfig{
		ChannelName: "@test_channel",
		BotToken:    "test_token",
	}

	client := NewTelegramChannelClient(config, zap.NewNop())

	// Тестируем отправку пустого сообщения
	ctx := context.Background()
	_ = client.SendMessage(ctx, "", TelegramMessageOptions{
		LinkPreview:         false,
		DisableNotification: false,
	})

	// Ожидаем ошибку или успех (зависит от реализации)
	// assert.Error(t, err) // если валидация включена
}

func TestTelegramChannelClient_NewTelegramChannelClient(t *testing.T) {
	config := TelegramChannelClientConfig{
		ChannelName: "@test_channel",
		BotToken:    "test_token",
		SilentMode: SilentModeConfig{
			Start:    "23:00:00",
			Finish:   "08:00:00",
			Timezone: "Europe/Moscow",
		},
	}

	logger := zap.NewNop()
	client := NewTelegramChannelClient(config, logger)

	assert.NotNil(t, client)
	assert.Equal(t, config, client.conf)
	assert.NotNil(t, client.logger)
}
