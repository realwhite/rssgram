package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
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

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func runMigrate() error {

	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed create sqlite3 instance: %w", err)
	}

	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to init iofs: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "sqlite3", instance)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to migrate: %w", err)
	}
	return nil
}

func main() {

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, _ := cfg.Build()
	defer logger.Sync()

	err := runMigrate()
	if err != nil {
		logger.Fatal("migrate failed", zap.Error(err))
	}

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

	itemsToSend, err := storage.GetItemsReadyToSend(ctx, 0)
	if err != nil {
		logger.Error("failed to get items to send", zap.Error(err))
	}

	metrics.ItemsReadyToSendCount.Set(float64(len(itemsToSend)))

	failedItems, err := storage.GetCountItemsSendFailed(ctx)
	if err != nil {
		logger.Error("failed to get failed items", zap.Error(err))
	}

	metrics.ItemsSentFailedCount.Set(float64(failedItems))

	logger.Debug(fmt.Sprintf("got %d items to send", len(itemsToSend)))

	for i := range itemsToSend {
		time.Sleep(1 * time.Second)
		logger.Debug(fmt.Sprintf("sending %s ...", itemsToSend[i].ID))

		pushCtx := context.WithValue(ctx, "item_id", itemsToSend[i].ID)
		isSuccess, err := tgOutput.Push(pushCtx, &itemsToSend[i])
		if err != nil {
			logger.Error("failed to send item", zap.Error(err))
			metrics.ItemsSentErrorCount.WithLabelValues(itemsToSend[i].FeedTitle).Inc()
			err = storage.IncrementItemFailedCounter(ctx, itemsToSend[i].ID)
			if err != nil {
				logger.Error("failed to increment item failed", zap.Error(err))
			}
			continue
		}
		if !isSuccess {
			logger.Error("failed to send item")
			metrics.ItemsSentErrorCount.WithLabelValues(itemsToSend[i].FeedTitle).Inc()
			err = storage.IncrementItemFailedCounter(ctx, itemsToSend[i].ID)
			if err != nil {
				logger.Error("failed to increment item failed", zap.Error(err))
			}
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
