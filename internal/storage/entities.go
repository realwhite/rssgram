package storage

import (
	"time"
)

type StoredFeed struct {
	URL         string
	LastChecked time.Time
	LastPosted  time.Time
}
