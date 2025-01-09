package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
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
	LinkPreview bool `json:"link_preview"`
}

type Photo struct {
	ChatID    string `json:"chat_id"`
	Photo     string `json:"photo"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

type TelegramChannelClient struct {
	conf       *Config
	httpClient http.Client
}

func (c *TelegramChannelClient) SendPhoto(msg string, photoUrl string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", c.conf.Telegram[0].Token)

	m := Photo{
		ChatID:    c.conf.Telegram[0].Name,
		Photo:     photoUrl,
		Caption:   msg,
		ParseMode: "HTML",
	}

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	reqDump, _ := httputil.DumpRequest(req, true)
	fmt.Printf("REQ:\n%s\n\n", string(reqDump))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	respDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("RESPONSE:\n%s", string(respDump))

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("return status code: %d", res.StatusCode)
	}

	return nil
}

func (c *TelegramChannelClient) SendMessage(msg string, options TelegramMessageOptions) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.conf.Telegram[0].Token)

	m := Message{
		ChatID:             c.conf.Telegram[0].Name,
		ParseMode:          "HTML",
		Text:               msg,
		LinkPreviewOptions: LinkPreviewOptions{IsDisabled: !options.LinkPreview},
	}

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	reqDump, _ := httputil.DumpRequest(req, true)
	fmt.Printf("REQ:\n%s\n\n", string(reqDump))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	respDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("RESPONSE:\n%s", string(respDump))

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("return status code: %d", res.StatusCode)
	}

	return nil
}

func NewTelegramChannelClient(conf *Config) *TelegramChannelClient {
	return &TelegramChannelClient{
		conf:       conf,
		httpClient: http.Client{},
	}
}
