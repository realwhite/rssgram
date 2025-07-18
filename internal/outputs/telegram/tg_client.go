package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"

	"go.uber.org/zap"
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrBadRequest      = errors.New("bad request")
)

type LinkPreviewOptions struct {
	IsDisabled bool `json:"is_disabled"`
}

type Message struct {
	ChatID               string             `json:"chat_id"`
	ParseMode            string             `json:"parse_mode"`
	Text                 string             `json:"text"`
	LinkPreviewOptions   LinkPreviewOptions `json:"link_preview_options"`
	DisableNotifications bool               `json:"disable_notification"`
}

type TelegramMessageOptions struct {
	LinkPreview         bool `json:"link_preview"`
	DisableNotification bool `json:"disable_notification"`
}

type Photo struct {
	ChatID              string `json:"chat_id"`
	Photo               string `json:"photo"`
	Caption             string `json:"caption"`
	ParseMode           string `json:"parse_mode"`
	DisableNotification bool   `json:"disable_notification"`
}

type SilentModeConfig struct {
	Start    string `json:"start"`
	Finish   string `json:"finish"`
	Timezone string `json:"timezone"`
}

type TelegramChannelClientConfig struct {
	ChannelName string           `yaml:"channel_name"`
	BotToken    string           `yaml:"bot_token"`
	SilentMode  SilentModeConfig `yaml:"silent_mode"`
}

type TelegramChannelClient struct {
	conf       TelegramChannelClientConfig
	httpClient http.Client

	logger *zap.Logger
}

func (c *TelegramChannelClient) SendPhoto(ctx context.Context, msg string, photoUrl string, disableNotification bool) error {

	ctxLogger := c.logger
	if v := ctx.Value("item_id"); v != nil {
		if itemID, ok := v.(string); ok {
			ctxLogger = c.logger.With(zap.String("item_id", itemID))
		}
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", c.conf.BotToken)

	m := Photo{
		ChatID:              c.conf.ChannelName,
		Photo:               photoUrl,
		Caption:             msg,
		ParseMode:           "HTML",
		DisableNotification: disableNotification,
	}

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return fmt.Errorf("failed to dump request: %w", err)
	}
	ctxLogger.Debug("request dump", zap.String("request", string(reqDump)))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	respDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return fmt.Errorf("failed to dump response: %w", err)
	}

	ctxLogger.Debug("response dump", zap.String("response", string(respDump)))

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("return status code: %d", res.StatusCode)
	}

	return nil
}

func (c *TelegramChannelClient) SendMessage(ctx context.Context, msg string, options TelegramMessageOptions) error {
	ctxLogger := c.logger
	if v := ctx.Value("item_id"); v != nil {
		if itemID, ok := v.(string); ok {
			ctxLogger = c.logger.With(zap.String("item_id", itemID))
		}
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.conf.BotToken)

	m := Message{
		ChatID:               c.conf.ChannelName,
		ParseMode:            "HTML",
		Text:                 msg,
		LinkPreviewOptions:   LinkPreviewOptions{IsDisabled: !options.LinkPreview},
		DisableNotifications: options.DisableNotification,
	}

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return fmt.Errorf("failed to dump request: %w", err)
	}
	ctxLogger.Debug("request dump", zap.String("request", string(reqDump)))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	respDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return fmt.Errorf("failed to dump response: %w", err)
	}
	ctxLogger.Debug("response dump", zap.String("response", string(respDump)))

	switch res.StatusCode {
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("return status code: %d", res.StatusCode)
	}
}

func NewTelegramChannelClient(conf TelegramChannelClientConfig, logger *zap.Logger) *TelegramChannelClient {
	return &TelegramChannelClient{
		conf:       conf,
		httpClient: http.Client{},
		logger:     logger,
	}
}
