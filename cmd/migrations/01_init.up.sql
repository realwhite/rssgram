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