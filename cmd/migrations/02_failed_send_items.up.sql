ALTER TABLE items ADD COLUMN failed_count INT NOT NULL DEFAULT 0;
ALTER TABLE items ADD COLUMN updated_at TEXT;