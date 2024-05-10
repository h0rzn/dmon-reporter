package store

// implemented by respective cache provider
type OfflineCache interface {
	// Initialize Cache by preparing and
	// connecting to the cache provider
	Init(config map[string]string) error

	// Push a dataset `CacheData` to the cache
	Push(Data) error

	// Get all stored data (for this session)
	Fetch() ([]Data, error)

	// Relase cache by dropping stored data
	// Maybe get all stored data (for session?)
	Drop() error

	// Unhook the cache, but don't drop data.
	// Useful for disconnecting and cleaning up
	// to retrieve cached data later
	Close()

	Run(chan Data)
}
