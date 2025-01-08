package main

import (
	"log"

	"rssgram/internal/feed"
	"rssgram/internal/storage/sqlite"

	"github.com/davecgh/go-spew/spew"
)

func main() {

	storage, err := sqlite.NewStorage()
	if err != nil {
		log.Fatal(err)
	}

	m := feed.NewManager(storage)

	_feed, err := m.GetFeed(feed.FeedConfig{
		Name:            "test_name",
		URL:             "https://news.ycombinator.com/rss",
		Key:             "",
		DescriptionType: "link",
		Tags:            []string{"test_tag"},
	})

	if err != nil {
		log.Fatal(err)
	}

	spew.Dump(_feed)

	m.EnrichFeedItems(_feed)

	spew.Dump(_feed)

}
