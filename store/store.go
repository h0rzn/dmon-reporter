package store

import (
	"time"
)

// implemented by respective cache provider
type OfflineCache interface {
	// Initialize Cache by preparing and
	// connecting to the cache provider
	Init(config map[string]string) error

	// Push a dataset `CacheData` to the cache
	Push(CacheData) error

	// Get all stored data (for this session)
	Fetch() ([]CacheData, error)

	// Relase cache by dropping stored data
	// Maybe get all stored data (for session?)
	Drop() error

	// Unhook the cache, but don't drop data.
	// Useful for disconnecting and cleaning up
	// to retrieve cached data later
	Close()
}

type CacheData interface {
	ID() string
	When() time.Time
	// Content() (json.RawMessage, error)
	Content() any
}
