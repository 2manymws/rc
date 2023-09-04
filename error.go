package rc

import (
	"errors"
)

// ErrCacheNotFound is returned when the cache is not found
var ErrCacheNotFound error = errors.New("cache not found")

// ErrNoCache is returned if not caching
var ErrNoCache error = errors.New("no cache")

// ErrCacheExpired is returned if the cache is expired
var ErrCacheExpired error = errors.New("cache expired")
