package main

import (
	"context"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
	"html"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	"rssgram/internal"
)

// читаем фиды из конфига
// синхронизируем их с базой (Storage)
// поднимаем из базы значения прошлой прошлой проверки и т.д. (FileStorage)
// проверяем для каких фидов уже подошел интервал (FeedFilter)
// если интервал еще не подошел - пропускаем (FeedFilter)
// если подошел: (FeedFilter)
// - получаем все фиды (FeedFiller)
// - определяем, есть ли там новые посты
// - собираем все посты из фидов с одинаковым KEY
// 		- дедуплицируем одинаковые посты - должен остаться только один пост прикрепленный к какому-то фиду
// - передаем на отправку (Sender)
//		- если тип LIST, то новые посты отправляем просто списком
//		- если SINGLE, то каждый пост отправляется отдельно

// StorageSync -> FeedFilter ->

func main() {

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, _ := cfg.Build()
	defer logger.Sync()

	//p := internal.SiteParser{}
	//fmt.Println(p.GetDescription("https://blog.transparency.dev/postgresql-support-for-certificate-transparency-logs-released"))
	//fmt.Println(p.GetDescription("https://www.crisesnotes.com/content/files/2023/12/NYFRB-2006.--Doomsday-Book--Searchable.pdf"))
	//return

	cnf, err := internal.ParseConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}

	tg := internal.NewTelegramChannelClient(cnf)

	storage, err := internal.NewSQLiteStorage()
	if err != nil {
		logger.Fatal(err.Error())
	}

	feeds, err := PrepareFeeds(cnf, storage, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	feeds, err = FilterFeedsByInterval(feeds, time.Now().UTC())
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Debug(fmt.Sprintf("%d feeds needs to update", len(feeds)))

	wg := &sync.WaitGroup{}

	for i := range feeds {
		wg.Add(1)

		go func() {
			ctxLogger := logger.With(zap.String("feed", feeds[i].FeedConfig.URL))
			defer func() {
				ctxLogger.Debug("finish processing")
				wg.Done()
			}()

			ctxLogger.Debug("start to processing feed")

			err = FillFeed(&feeds[i])
			if err != nil {
				ctxLogger.Error("failed to fill feed", zap.Error(err))
				return
			}

			newItems, newLastPosted := FilterFeedItemsByLastPosted(feeds[i].Feed.Items, feeds[i].LastPosted)
			feeds[i].Feed.Items = newItems

			ctxLogger.Debug(fmt.Sprintf("found %d new items", len(newItems)))

			FillFeedItems(&feeds[i])

			if !feeds[i].IsNew {
				logger.Debug("start to send messages")
				var sendErr error
				if feeds[i].Type == "list" || (feeds[i].Type == "" && len(newItems) > 30) {
					sendErr = SendFeedList(&feeds[i], tg)
				} else {
					sendErr = SendFeedPost(&feeds[i], tg)
				}

				if sendErr != nil {
					ctxLogger.Error("failed to send feed", zap.Error(sendErr))
					return
				}

			}

			err = storage.UpsertFeed(feeds[i].FeedConfig.URL, time.Now(), newLastPosted)
			if err != nil {
				ctxLogger.Error("failed to update feed in db", zap.Error(err))
			}
		}()
	}

	wg.Wait()

	return
}

func PrepareFeeds(cnf *internal.Config, storage *internal.SQLiteStorage, log *zap.Logger) ([]internal.CommonFeed, error) {
	var feeds []internal.CommonFeed
	storedFeeds, err := storage.GetAllFeeds()
	if err != nil {
		return nil, fmt.Errorf("error getting all feeds: %w", err)
	}

	log.Debug(fmt.Sprintf("Found %d feeds in config", len(cnf.Feeds)))
	log.Debug(fmt.Sprintf("Found %d feeds in database", len(storedFeeds)))

	storedFeedMap := make(map[string]*internal.StoredFeed)
	for _, feed := range storedFeeds {
		storedFeedMap[feed.URL] = &feed
	}

	configFeedMap := make(map[string]*internal.FeedConfig)
	for _, cf := range cnf.Feeds {
		configFeedMap[cf.URL] = &cf

		if _, ok := storedFeedMap[cf.URL]; !ok {
			log.Debug(fmt.Sprintf("Adding feed %s to database", cf.URL))
			err = storage.UpsertFeed(cf.URL, time.Time{}, time.Time{})
			if err != nil {
				return nil, fmt.Errorf("error upserting feed %s: %w", cf.URL, err)
			}
			storedFeedMap[cf.URL] = &internal.StoredFeed{URL: cf.URL, LastChecked: time.Time{}, LastPosted: time.Time{}, IsNew: true}
			log.Debug("success")
		}
	}

	for _, feed := range storedFeeds {
		if _, ok := configFeedMap[feed.URL]; !ok {
			log.Debug(fmt.Sprintf("Delete feed %s from database", feed.URL))
			err = storage.DeleteFeed(feed.URL)
			delete(storedFeedMap, feed.URL)
			if err != nil {
				return nil, fmt.Errorf("error deleting feed %s: %w", feed.URL, err)
			}
			log.Debug("success")
		}
	}

	for _, cf := range cnf.Feeds {
		feeds = append(feeds, internal.CommonFeed{
			StoredFeed: *storedFeedMap[cf.URL],
			FeedConfig: cf,
		})
	}

	return feeds, nil
}

func FilterFeedsByInterval(feeds []internal.CommonFeed, refTime time.Time) ([]internal.CommonFeed, error) {
	var resultFeeds []internal.CommonFeed
	for _, feed := range feeds {
		interval, err := feed.GetInterval()
		if err != nil {
			return nil, fmt.Errorf("failed to get feed interval: %w", err)
		}
		if feed.LastChecked.Add(interval).Before(refTime) {
			resultFeeds = append(resultFeeds, feed)
		}
	}

	return resultFeeds, nil
}

func FilterFeedItemsByLastPosted(items []*gofeed.Item, lastPosted time.Time) ([]*gofeed.Item, time.Time) {
	var resultItems []*gofeed.Item
	maxLastPosted := lastPosted

	for _, item := range items {
		if item.PublishedParsed.Compare(lastPosted) > 0 {
			resultItems = append(resultItems, item)
			if item.PublishedParsed.Compare(maxLastPosted) > 0 {
				maxLastPosted = *item.PublishedParsed
			}
		}
	}

	return resultItems, maxLastPosted
}

func FillFeed(f *internal.CommonFeed) error {
	fp := gofeed.NewParser()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	respFeed, err := fp.ParseURLWithContext(f.FeedConfig.URL, ctx)
	cancel() // to defer?
	if err != nil {
		return err
	}

	f.Feed = respFeed
	return nil
}

func FillFeedItems(f *internal.CommonFeed) {
	if f.IsNew {
		// для только что добавленых фидов нам не надо обрабатывать айтемы
		return
	}
	if f.FeedConfig.DescriptionType == "link" {
		for i := range f.Feed.Items {
			url := f.Feed.Items[i].Link
			p := internal.SiteParser{}
			siteDescription, err := p.GetDescription(url)
			if err != nil {
				continue
			}

			if siteDescription.Description != "" {
				f.Feed.Items[i].Description = siteDescription.Description
			} else if siteDescription.Title != "" {
				f.Feed.Items[i].Description = siteDescription.Title
			}

			if siteDescription.Image != "" {
				im := gofeed.Image{
					URL:   siteDescription.Image,
					Title: "",
				}

				f.Feed.Items[i].Image = &im
			}
		}
	}
}

func SendFeedList(f *internal.CommonFeed, tg *internal.TelegramChannelClient) error {

	s := MessageSlicer{}
	messages, _ := s.SliceFeed(f.Feed)
	for _, m := range messages {
		err := tg.SendMessage(m, internal.TelegramMessageOptions{LinkPreview: false})
		if err != nil {
			log.Fatalf("error sending message to telegram channel: %v", err)
		}
	}

	return nil
}

func SendFeedPost(f *internal.CommonFeed, tg *internal.TelegramChannelClient) error {
	for _, item := range f.Feed.Items {
		time.Sleep(5 * time.Second)
		feedTitle := fmt.Sprintf("<b>[%s]</b>", f.Name)
		itemTitle := fmt.Sprintf("<a href=\"%s\">%s</a>", item.Link, html.EscapeString(item.Title))

		p := bluemonday.StripTagsPolicy()

		description := fmt.Sprintf("%s", p.Sanitize(item.Description))

		var sendErr error
		if item.Image == nil || item.Image.URL == "" {
			msg := fmt.Sprintf("%s\n\n%s\n\n%s", feedTitle, itemTitle, fmt.Sprintf("<blockquote>%s</blockquote>", description))
			sendErr = tg.SendMessage(msg, internal.TelegramMessageOptions{LinkPreview: false})
		} else {
			shortDescription := ellipsisString(description, 800)
			msg := fmt.Sprintf("%s\n\n%s\n\n%s", feedTitle, itemTitle, fmt.Sprintf("<blockquote>%s</blockquote>", shortDescription))
			sendErr = tg.SendPhoto(msg, item.Image.URL)
		}

		if sendErr != nil {
			log.Fatalf("error sending message to telegram channel: %v", sendErr)
		}
	}

	return nil
}

func ellipsisString(s string, max int) string {
	if max > len(s) {
		return s
	}
	return s[:strings.LastIndexAny(s[:max], " .,:;-")] + "..."
}

// move to telegram output
var msgThreshhold = 4000

type MessageSlicer struct{}

func (s MessageSlicer) SliceFeed(f *gofeed.Feed) ([]string, error) {
	var msgSlices []string

	title := fmt.Sprintf("<b>[%s]</b>", f.Title)
	entries := []string{}
	size := 0
	size2 := 0

	for _, i := range f.Items {
		itemLink := fmt.Sprintf("<a href=\"%s\">%s</a>", i.Link, html.EscapeString(i.Title))
		if size+len(i.Title) > msgThreshhold || size2+len(i.Link) > 9000 {
			msgSlices = append(msgSlices, fmt.Sprintf("%s\n\n%s", title, strings.Join(entries, "\n")))
			entries = []string{}
			size = 0
			size2 = 0
		}

		entries = append(entries, itemLink)
		size += len(i.Title)
		size2 += len(i.Link)
	}

	if len(entries) > 0 {
		msgSlices = append(msgSlices, fmt.Sprintf("%s\n\n%s", title, strings.Join(entries, "\n")))
	}

	//fmt.Printf("\n\n\n %s \n\n\n", msgSlices)

	return msgSlices, nil
}
