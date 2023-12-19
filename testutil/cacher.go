package testutil

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	_ Cacher = &AllCache{}
	_ Cacher = &GetOnlyCache{}
)

// errCacheNotFound is returned when the cache is not found
var errCacheNotFound error = errors.New("cache not found")

// errNoCache is returned if not caching
var errNoCache error = errors.New("no cache")

type Cacher interface {
	Load(req *http.Request) (cachedReq *http.Request, cachedRes *http.Response, err error)
	Store(req *http.Request, res *http.Response, expires time.Time) error
	Hit() int
}

type AllCache struct {
	t   testing.TB
	m   map[string]*cachedReqRes
	dir string
	hit int
	mu  sync.Mutex
}

type GetOnlyCache struct {
	t   testing.TB
	m   map[string]*cachedReqRes
	dir string
	hit int
	mu  sync.Mutex
}

func NewAllCache(t testing.TB) *AllCache {
	t.Helper()
	return &AllCache{
		t:   t,
		m:   map[string]*cachedReqRes{},
		dir: t.TempDir(),
	}
}

func (c *AllCache) Load(req *http.Request) (*http.Request, *http.Response, error) {
	c.t.Helper()
	key := reqToKey(req)
	c.mu.Lock()
	cc, ok := c.m[key]
	c.mu.Unlock()
	if !ok {
		return nil, nil, errCacheNotFound
	}
	cachedReq, cachedRes, err := decodeReqRes(cc)
	if err != nil {
		return nil, nil, err
	}
	cachedRes.Header.Set("X-Cache", "HIT")
	c.hit++
	return cachedReq, cachedRes, nil
}

func (c *AllCache) Store(req *http.Request, res *http.Response, expires time.Time) error {
	c.t.Helper()
	key := reqToKey(req)
	cc, err := encodeReqRes(req, res)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.m[key] = cc
	c.mu.Unlock()
	return nil
}

func (c *AllCache) Hit() int {
	return c.hit
}

func NewGetOnlyCache(t testing.TB) *GetOnlyCache {
	t.Helper()
	return &GetOnlyCache{
		t:   t,
		m:   map[string]*cachedReqRes{},
		dir: t.TempDir(),
	}
}

func (c *GetOnlyCache) Load(req *http.Request) (*http.Request, *http.Response, error) {
	c.t.Helper()
	if req.Method != http.MethodGet {
		return nil, nil, errNoCache
	}
	key := reqToKey(req)
	c.mu.Lock()
	cc, ok := c.m[key]
	c.mu.Unlock()
	if !ok {
		return nil, nil, errCacheNotFound
	}
	cachedReq, cachedRes, err := decodeReqRes(cc)
	if err != nil {
		return nil, nil, err
	}
	cachedRes.Header.Set("X-Cache", "HIT")
	c.hit++
	return cachedReq, cachedRes, nil
}

func (c *GetOnlyCache) Store(req *http.Request, res *http.Response, expires time.Time) error {
	c.t.Helper()
	if req.Method != http.MethodGet {
		return errNoCache
	}
	key := reqToKey(req)
	cc, err := encodeReqRes(req, res)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.m[key] = cc
	c.mu.Unlock()
	return nil
}

func (c *GetOnlyCache) Hit() int {
	c.t.Helper()
	return c.hit
}

func reqToKey(req *http.Request) string {
	const sep = "|"
	seed := req.Method + sep + req.Host + sep + req.URL.Host + sep + req.URL.Path + sep + req.URL.RawQuery
	sha1 := sha1.New()
	_, _ = io.WriteString(sha1, strings.ToLower(seed)) //nostyle:handlerrors
	return hex.EncodeToString(sha1.Sum(nil))
}
