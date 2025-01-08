package internal

import (
	"crypto/sha256"
	"time"

	"github.com/mmcdole/gofeed"
)

const defaultInterval = time.Second

type StoredFeed struct {
	URL         string
	LastChecked time.Time
	LastPosted  time.Time
	IsNew       bool
}

type CommonFeed struct {
	FeedConfig
	StoredFeed

	Feed *gofeed.Feed
}

func (f *CommonFeed) GetKey() string {
	if f.Key == "" {
		return string(sha256.New().Sum([]byte(f.FeedConfig.URL)))
	}

	return f.Key
}

func (f *CommonFeed) GetInterval() (time.Duration, error) {
	if f.Interval == "" {
		return defaultInterval, nil
	}

	interval, err := time.ParseDuration(f.Interval)
	if err != nil {
		return defaultInterval, err
	}

	return interval, nil
}

func NewCommonFeed(fc *FeedConfig, sf *StoredFeed) CommonFeed {
	return CommonFeed{
		FeedConfig: *fc,
		StoredFeed: *sf,
	}
}
