package telegram

import (
	"context"
	"fmt"
	"strings"
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
	client TelegramClient

	config     TelegramChannelOutputConfig
	enableTags bool
}

func (o *TelegramChannelOutput) IsSilentMode(startTimeStr, finishTimeStr, tzStr string, refTime time.Time) (bool, error) {

	if startTimeStr == "" || finishTimeStr == "" {
		return false, nil
	}

	startTime, err := time.Parse(time.TimeOnly, startTimeStr)
	if err != nil {
		return false, fmt.Errorf("error parsing start time: %w", err)
	}

	finishTime, err := time.Parse(time.TimeOnly, finishTimeStr)
	if err != nil {
		return false, fmt.Errorf("error parsing finish time: %w", err)
	}

	loc, err := time.LoadLocation(tzStr)
	if err != nil {
		return false, fmt.Errorf("error loading location: %w", err)
	}

	refTime = refTime.In(loc)
	onlyRefTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), refTime.Hour(), refTime.Minute(), refTime.Second(), refTime.Nanosecond(), loc)

	if finishTime.After(startTime) {
		// one day
		if onlyRefTime.After(startTime) && onlyRefTime.Before(finishTime) {
			return true, nil
		}
		return false, nil
	} else if finishTime.Before(startTime) {
		// crossday
		if onlyRefTime.After(finishTime) && onlyRefTime.Before(startTime) {
			return false, nil
		}
		return true, nil
	}

	return true, nil
}

func (o *TelegramChannelOutput) Push(ctx context.Context, item *feed.FeedItem) (bool, error) {

	disableNotification, err := o.IsSilentMode(
		o.config.SilentMode.Start,
		o.config.SilentMode.Finish,
		o.config.SilentMode.Timezone,
		time.Now(),
	)
	if err != nil {
		return false, fmt.Errorf("error checking silent mode: %w", err)
	}

	feedTitle := fmt.Sprintf("<b>[%s]</b>", item.FeedTitle)
	itemTitle := fmt.Sprintf("<a href=\"%s\">%s</a>", item.Link, html.EscapeString(item.Title))

	p := bluemonday.StripTagsPolicy()

	description := fmt.Sprintf("%s", p.Sanitize(item.Description))
	shortDescription := utils.EllipsisString(description, 800)

	msg := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		feedTitle,
		itemTitle,
		fmt.Sprintf("<blockquote>%s</blockquote>", shortDescription),
	)

	// Добавляем теги, если включено
	if o.enableTags && len(item.Tags) > 0 {
		tags := ""
		for _, tag := range item.Tags {
			tag = strings.ReplaceAll(tag, " ", "_")
			tags += "#" + tag + " "
		}
		tags = tags[:len(tags)-1] // убрать последний пробел
		msg += "\n\n" + tags
	}

	var sendErr error

	if item.ImageURL == "" {
		sendErr = o.client.SendMessage(
			ctx,
			msg,
			TelegramMessageOptions{LinkPreview: false, DisableNotification: disableNotification},
		)
	} else {
		sendErr = o.client.SendPhoto(ctx, msg, item.ImageURL, disableNotification)
	}

	if sendErr != nil {
		return false, sendErr
	}

	return true, nil
}

func NewTelegramChannelOutput(conf TelegramChannelOutputConfig, logger *zap.Logger, enableTags bool) *TelegramChannelOutput {
	return &TelegramChannelOutput{
		config:     conf,
		client:     TelegramClient(NewTelegramChannelClient(conf.TelegramChannelClientConfig, logger)),
		enableTags: enableTags,
	}
}

// Интерфейс для клиента Telegram

type TelegramClient interface {
	SendMessage(ctx context.Context, msg string, options TelegramMessageOptions) error
	SendPhoto(ctx context.Context, msg, photoUrl string, disableNotification bool) error
}
