package interfaces

import "github.com/poteto0/takibi/constants"

// TakibiOption holds framework-level configuration for NewWithOption.
// Zero values fall back to their documented defaults.
type TakibiOption struct {
	// MaxBodyBytes limits request body size decoded by Unmarshall.
	// 0 uses the default (constants.DefaultMaxBodyBytes = 10 MiB).
	MaxBodyBytes int64
}

var DefaultTakibiOption = TakibiOption{
	MaxBodyBytes: constants.DefaultMaxBodyBytes,
}
