package internal

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type SQLiteStorage struct {
	db *sql.DB
}

func (s *SQLiteStorage) GetAllFeeds() ([]StoredFeed, error) {
	stmt := "SELECT url, last_checked, last_post FROM feeds"
	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all feeds: %w", err)
	}
	defer rows.Close()
	var feeds []StoredFeed
	for rows.Next() {

		var url, lastChecked, lastPosted string

		err = rows.Scan(&url, &lastChecked, &lastPosted)
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

		feed := StoredFeed{
			URL:         url,
			LastChecked: parsedLastChecked,
			LastPosted:  parsedLastPosted,
		}

		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func (s *SQLiteStorage) DeleteFeed(url string) error {
	stmt := "DELETE FROM feeds WHERE url=?"
	_, err := s.db.Exec(stmt, url)
	return err
}

func (s *SQLiteStorage) UpsertFeed(url string, lastChecked, lastPost time.Time) error {
	stmt := "INSERT INTO feeds (url, last_checked, last_post) VALUES (?, ?, ?) ON CONFLICT(url) DO UPDATE SET last_checked=excluded.last_checked, last_post=excluded.last_post"
	_, err := s.db.Exec(stmt, url, lastChecked.UTC().Format(time.DateTime), lastPost.UTC().Format(time.DateTime))
	return err
}

func (s *SQLiteStorage) Init() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS feeds (
		    url TEXT NOT NULL PRIMARY KEY, 
		    last_checked TEXT NOT NULL, 
		    last_post TEXT NOT NULL
		);
`
	_, err := s.db.Exec(stmt)
	return err
}

func NewSQLiteStorage() (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite storage: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	err = storage.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite storage: %w", err)
	}
	return storage, nil
}
