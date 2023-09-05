package rc

import (
	"errors"
)

// ErrCacheNotFound is returned when the cache is not found
var ErrCacheNotFound error = errors.New("cache not found")

// ErrDoNotUseCache is returned if should not use cache
var ErrDoNotUseCache error = errors.New("do not use cache")

// ErrCacheExpired is returned if the cache is expired
var ErrCacheExpired error = errors.New("cache expired")
