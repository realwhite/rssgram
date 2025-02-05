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

	crossNight := false

	if finishTime.Hour() <= startTime.Hour() {
		crossNight = true
	}

	if !crossNight {
		startDateTime := time.Date(refTime.Year(), refTime.Month(), refTime.Day(), startTime.Hour(), startTime.Minute(), 0, 0, loc)
		finishDateTime := time.Date(refTime.Year(), refTime.Month(), refTime.Day(), finishTime.Hour(), finishTime.Minute(), 0, 0, loc)

		fmt.Println(refTime.After(startDateTime))
		fmt.Println(refTime.Before(finishDateTime))

		if startDateTime.Before(refTime) && finishDateTime.After(refTime) {
			return true, nil
		}
		return false, nil
	}

	if refTime.Hour() < finishTime.Hour() && refTime.Minute() <= finishTime.Minute() {
		return true, nil
	}

	if refTime.Hour() >= startTime.Hour() && refTime.Minute() >= startTime.Minute() {
		return true, nil
	}

	return false, nil

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

func NewTelegramChannelOutput(conf TelegramChannelOutputConfig, logger *zap.Logger) *TelegramChannelOutput {
	return &TelegramChannelOutput{
		config: conf,
		client: NewTelegramChannelClient(conf.TelegramChannelClientConfig, logger),
	}
}
