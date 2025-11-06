package rc

import (
	"errors"
)

// ErrCacheNotFound is returned when the cache is not found.
var ErrCacheNotFound error = errors.New("cache not found")

// ErrShouldNotUseCache is returned if should not use cache.
var ErrShouldNotUseCache error = errors.New("should not use cache")

// ErrCacheExpired is returned if the cache is expired.
var ErrCacheExpired error = errors.New("cache expired")
