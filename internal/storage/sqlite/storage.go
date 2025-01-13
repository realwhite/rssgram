package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"rssgram/internal/feed"
	"rssgram/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) DeleteFeed(ctx context.Context, url string) error {
	stmt := "DELETE FROM feeds WHERE url=?"
	_, err := s.db.Exec(stmt, url)
	return err
}

func (s *Storage) UpsertFeed(ctx context.Context, url string, lastChecked, lastPost time.Time) error {
	stmt := "INSERT INTO feeds (url, last_checked, last_post) VALUES (?, ?, ?) ON CONFLICT(url) DO UPDATE SET last_checked=excluded.last_checked, last_post=excluded.last_post"
	_, err := s.db.Exec(stmt, url, lastChecked.UTC().Format(time.DateTime), lastPost.UTC().Format(time.DateTime))
	return err
}

func (s *Storage) GetFeedByURL(ctx context.Context, url string) (*storage.StoredFeed, error) {
	stmt := "SELECT last_checked, last_post FROM feeds WHERE url=?"
	rows, err := s.db.Query(stmt, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var lastChecked, lastPosted string

		err = rows.Scan(&lastChecked, &lastPosted)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all feeds: %w", err)
		}

		parsedLastChecked, err := time.Parse(time.DateTime, lastChecked)
		if err != nil {
			return nil, fmt.Errorf("failed to convert last_cheked (%s): %w", url, err)
		}

		parsedLastPosted, err := time.Parse(time.DateTime, lastPosted)
		if err != nil {
			return nil, fmt.Errorf("failed to convert last_posted (%s): %w", url, err)
		}

		_feed := storage.StoredFeed{
			URL:         url,
			LastChecked: parsedLastChecked,
			LastPosted:  parsedLastPosted,
		}

		return &_feed, nil
	}

	return nil, nil
}

func (s *Storage) InsertItem(ctx context.Context, item *feed.FeedItem) error {
	itemMetaJSON, err := item.GetMetadataJson()
	if err != nil {
		return fmt.Errorf("failed to marshal item metadata: %w", err)
	}

	itemTagsJSON, err := item.GetTagsJson()
	if err != nil {
		return fmt.Errorf("failed to marshal item tags: %w", err)
	}

	stmt := `
	INSERT INTO items (id, feed_title, title, link, description, image_url, tags, metadata, published_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO NOTHING
`
	_, err = s.db.Exec(stmt,
		item.ID,
		item.FeedTitle,
		item.Title,
		item.Link,
		item.Description,
		item.ImageURL,
		itemTagsJSON,
		itemMetaJSON,
		item.PublishedAt.Format(time.DateTime),
	)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}
	return nil
}

func (s *Storage) GetItemsReadyToSend(ctx context.Context, limit int) ([]feed.FeedItem, error) {
	stmt := "SELECT id, title, feed_title, title, link, image_url, description, published_at, tags, metadata  FROM items where is_sent = 0 order by published_at limit ?"
	rows, err := s.db.Query(stmt, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ready items: %w", err)
	}
	defer rows.Close()
	var items []feed.FeedItem
	for rows.Next() {
		item := feed.FeedItem{}

		var publishedAt, tmpTags, tmpMeta string

		err = rows.Scan(&item.ID, &item.Title, &item.FeedTitle, &item.Title, &item.Link, &item.ImageURL, &item.Description, &publishedAt, &tmpTags, &tmpMeta)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all feeds: %w", err)
		}

		parsedPublishedAt, err := time.Parse(time.DateTime, publishedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert published_at (%s): %w", item, err)
		}
		item.PublishedAt = &parsedPublishedAt

		items = append(items, item)
	}

	return items, nil
}

func (s *Storage) SetItemIsSent(ctx context.Context, itemID string) error {
	stmt := `UPDATE items SET is_sent = 1, sent_at = ? WHERE id=?`
	_, err := s.db.Exec(stmt, time.Now().UTC().Format(time.DateTime), itemID)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	return nil
}

func (s *Storage) init() error {
	stmt := `
	CREATE TABLE IF NOT EXISTS feeds (
		url TEXT NOT NULL PRIMARY KEY, 
		last_checked TEXT NOT NULL, 
		last_post TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS items (
		id TEXT NOT NULL PRIMARY KEY,
		feed_title TEXT NOT NULL,
		title TEXT NOT NULL,
		link TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL,
		image_url TEXT,
		tags TEXT NOT NULL DEFAULT '',
		metadata TEXT NOT NULL DEFAULT '{}',
		published_at TEXT NOT NULL,
		is_sent BOOLEAN NOT NULL DEFAULT '0',
		sent_at TEXT
	);

`
	_, err := s.db.Exec(stmt)
	return err
}

func NewStorage() (*Storage, error) {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite storage: %w", err)
	}

	storage := &Storage{db: db}
	err = storage.init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite storage: %w", err)
	}
	return storage, nil
}
