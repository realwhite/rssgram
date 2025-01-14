package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rssgram/internal"
	"rssgram/internal/feed"
	"rssgram/internal/metrics"
	"rssgram/internal/outputs/telegram"
	"rssgram/internal/storage/sqlite"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, _ := cfg.Build()
	defer logger.Sync()

	cnf, err := internal.ParseConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}

	storage, err := sqlite.NewStorage()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	go feedGetter(ctx, cnf, storage, logger.With(zap.String("module", "feed_manager")))
	go itemSender(ctx, cnf, storage, logger.With(zap.String("module", "sender")))
	go metricHandler(ctx, cnf, logger.With(zap.String("module", "metric_handler")))

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	logger.Info(fmt.Sprintf("received signal %s", s.String()))

}

func feedGetter(ctx context.Context, cnf *internal.Config, storage *sqlite.Storage, logger *zap.Logger) {
	ticker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			ticker.Stop()
			_feedGetter(ctx, cnf, storage, logger)
			ticker.Reset(10 * time.Second)
		}

	}
}

func _feedGetter(ctx context.Context, cnf *internal.Config, storage *sqlite.Storage, logger *zap.Logger) {
	m := feed.NewManager(storage)

	metrics.FeedsCount.Set(float64(len(cnf.Feeds)))

	for _, f := range cnf.Feeds {
		// временный перегон из старого ConfigFeed.
		fc := feed.FeedConfig{
			Name:            f.Name,
			URL:             f.URL,
			Key:             f.Key,
			DescriptionType: f.DescriptionType,
		}

		err := m.ProcessFeed(context.Background(), fc, logger)
		if err != nil {
			logger.Error("failed to process feed", zap.String("url", f.URL), zap.Error(err))
			continue
		}
	}
}

func itemSender(ctx context.Context, cnf *internal.Config, storage *sqlite.Storage, logger *zap.Logger) {
	ticker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			ticker.Stop()
			_itemSender(ctx, cnf, storage, logger)
			ticker.Reset(10 * time.Second)
		}

	}
}

func _itemSender(ctx context.Context, cnf *internal.Config, storage *sqlite.Storage, logger *zap.Logger) {
	tgOutput := telegram.NewTelegramChannelOutput(
		cnf.Telegram,
		logger,
	)

	countItemsToSend, err := storage.GetCountItemsReadyToSend(ctx)
	if err != nil {
		logger.Error("failed to get count items to send", zap.Error(err))
	}

	metrics.ItemsReadyToSendCount.Add(float64(countItemsToSend))

	itemsToSend, err := storage.GetItemsReadyToSend(ctx, 1)
	if err != nil {
		logger.Error("failed to get items to send", zap.Error(err))
	}

	logger.Debug(fmt.Sprintf("got %d items to send", len(itemsToSend)))

	for i := range itemsToSend {
		logger.Debug(fmt.Sprintf("sending %s ...", itemsToSend[i].ID))

		pushCtx := context.WithValue(ctx, "item_id", itemsToSend[i].ID)
		isSuccess, err := tgOutput.Push(pushCtx, &itemsToSend[i])
		if err != nil {
			logger.Error("failed to send item", zap.Error(err))
			metrics.ItemsSentErrorCount.WithLabelValues(itemsToSend[i].FeedTitle).Inc()
			continue
		}
		if !isSuccess {
			logger.Error("failed to send item")
			metrics.ItemsSentErrorCount.WithLabelValues(itemsToSend[i].FeedTitle).Inc()
			continue
		}

		err = storage.SetItemIsSent(ctx, itemsToSend[i].ID)
		if err != nil {
			logger.Error("failed to set is_sent for item", zap.Error(err))
			continue
		}
		logger.Debug("sent")

		metrics.ItemsSentSuccessCount.WithLabelValues(itemsToSend[i].FeedTitle).Inc()

	}

}

func metricHandler(ctx context.Context, cnf *internal.Config, logger *zap.Logger) {

	http.Handle("/metrics", promhttp.Handler())
	logger.Info("start to serve /metrics on 2222 port")
	http.ListenAndServe(":2222", nil)
}
