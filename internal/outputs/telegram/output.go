package telegram

import (
	"context"
	"fmt"
	"time"

	"rssgram/internal/feed"
	"rssgram/internal/utils"

	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

type TelegramChannelOutputConfig struct {
	TelegramChannelClientConfig `yaml:",inline"`
}

type TelegramChannelOutput struct {
	client *TelegramChannelClient

	config TelegramChannelOutputConfig
}

func (o *TelegramChannelOutput) isSilentMode() (bool, error) {
	if o.config.SilentMode.Start != "" && o.config.SilentMode.Finish != "" {
		startTime, err := time.Parse(time.TimeOnly, o.config.SilentMode.Start)
		if err != nil {
			return false, fmt.Errorf("error parsing start time: %w", err)
		}

		finishTime, err := time.Parse(time.TimeOnly, o.config.SilentMode.Finish)
		if err != nil {
			return false, fmt.Errorf("error parsing finish time: %w", err)
		}

		loc, err := time.LoadLocation(o.config.SilentMode.Timezone)
		if err != nil {
			return false, fmt.Errorf("error loading location: %w", err)
		}

		now := time.Now().In(loc)
		daysDelta := 0

		startDateTime := time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), 0, 0, loc)

		if finishTime.Hour() < startTime.Hour() {
			daysDelta = 1
		}

		finishDateTime := time.Date(now.Year(), now.Month(), now.Day()+daysDelta, finishTime.Hour(), finishTime.Minute(), 0, 0, loc)

		if now.After(startDateTime) && now.Before(finishDateTime) {
			return true, nil
		}
	}

	return false, nil

}

func (o *TelegramChannelOutput) Push(ctx context.Context, item *feed.FeedItem) (bool, error) {

	disableNotification, err := o.isSilentMode()
	if err != nil {
		return false, fmt.Errorf("error checking silent mode: %w", err)
	}

	feedTitle := fmt.Sprintf("<b>[%s]</b>", item.FeedTitle)
	itemTitle := fmt.Sprintf("<a href=\"%s\">%s</a>", item.Link, html.EscapeString(item.Title))

	p := bluemonday.StripTagsPolicy()

	description := fmt.Sprintf("%s", p.Sanitize(item.Description))

	var sendErr error
	if item.ImageURL == "" {
		shortDescription := utils.EllipsisString(description, 800)
		msg := fmt.Sprintf("%s\n\n%s\n\n%s", feedTitle, itemTitle, fmt.Sprintf("<blockquote>%s</blockquote>", shortDescription))
		sendErr = o.client.SendMessage(ctx, msg, TelegramMessageOptions{LinkPreview: false, DisableNotification: disableNotification})
	} else {
		shortDescription := utils.EllipsisString(description, 800)
		msg := fmt.Sprintf("%s\n\n%s\n\n%s", feedTitle, itemTitle, fmt.Sprintf("<blockquote>%s</blockquote>", shortDescription))
		sendErr = o.client.SendPhoto(ctx, msg, item.ImageURL, disableNotification)
	}

	if sendErr != nil {
		return false, sendErr
	}

	return true, nil
}

func NewTelegramChannelOutput(conf TelegramChannelOutputConfig, logger *zap.Logger) *TelegramChannelOutput {
	return &TelegramChannelOutput{
		config: conf,
		client: NewTelegramChannelClient(conf.TelegramChannelClientConfig, logger),
	}
}
