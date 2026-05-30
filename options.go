package takibi

import "github.com/poteto0/takibi/constants"

// TakibiOption holds framework-level configuration passed to NewWithOption.
// Zero values fall back to their documented defaults.
type TakibiOption struct {
	// MaxBodyBytes limits the size of request bodies decoded by Unmarshall.
	// 0 means use the default (constants.DefaultMaxBodyBytes = 10 MiB).
	MaxBodyBytes int64
}

func defaultTakibiOption() TakibiOption {
	return TakibiOption{
		MaxBodyBytes: constants.DefaultMaxBodyBytes,
	}
}
