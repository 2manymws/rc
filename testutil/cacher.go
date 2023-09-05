package testutil

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/k1LoW/rcutil"
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
	Name() string
	Load(req *http.Request) (res *http.Response, err error)
	Store(req *http.Request, res *http.Response) error
	Hit() int
}

type AllCache struct {
	t   testing.TB
	m   map[string]string
	dir string
	hit int
	mu  sync.Mutex
}

type GetOnlyCache struct {
	t   testing.TB
	m   map[string]string
	dir string
	hit int
	mu  sync.Mutex
}

func NewAllCache(t testing.TB) *AllCache {
	t.Helper()
	return &AllCache{
		t:   t,
		m:   map[string]string{},
		dir: t.TempDir(),
	}
}

func (c *AllCache) Name() string {
	c.t.Helper()
	return "all"
}

func (c *AllCache) Load(req *http.Request) (res *http.Response, err error) {
	c.t.Helper()
	c.mu.Lock()
	p, ok := c.m[req.URL.Path]
	c.mu.Unlock()
	if !ok {
		return nil, errCacheNotFound
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, errCacheNotFound
	}
	res, err = rcutil.BytesToResponse(b)
	if err != nil {
		return nil, err
	}
	res.Header.Set("X-Cache", "HIT")
	c.hit++
	return res, nil
}

func (c *AllCache) Store(req *http.Request, res *http.Response) error {
	c.t.Helper()
	b, err := rcutil.ResponseToBytes(res)
	if err != nil {
		return err
	}
	p := filepath.Join(c.dir, base64.URLEncoding.EncodeToString([]byte(req.URL.Path)))
	if err := os.WriteFile(p, b, os.ModePerm); err != nil {
		return err
	}
	c.mu.Lock()
	c.m[req.URL.Path] = p
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
		m:   map[string]string{},
		dir: t.TempDir(),
	}
}

func (c *GetOnlyCache) Name() string {
	c.t.Helper()
	return "get-only"
}

func (c *GetOnlyCache) Load(req *http.Request) (res *http.Response, err error) {
	c.t.Helper()
	if req.Method != http.MethodGet {
		return nil, errNoCache
	}
	c.mu.Lock()
	p, ok := c.m[req.URL.Path]
	c.mu.Unlock()
	if !ok {
		return nil, errCacheNotFound
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, errCacheNotFound
	}
	res, err = rcutil.BytesToResponse(b)
	if err != nil {
		return nil, err
	}
	res.Header.Set("X-Cache", "HIT")
	c.hit++
	return res, nil
}

func (c *GetOnlyCache) Store(req *http.Request, res *http.Response) error {
	c.t.Helper()
	if req.Method != http.MethodGet {
		return errNoCache
	}
	b, err := rcutil.ResponseToBytes(res)
	if err != nil {
		return err
	}
	p := filepath.Join(c.dir, base64.URLEncoding.EncodeToString([]byte(req.URL.Path)))
	if err := os.WriteFile(p, b, os.ModePerm); err != nil {
		return err
	}
	c.mu.Lock()
	c.m[req.URL.Path] = p
	c.mu.Unlock()
	return nil
}

func (c *GetOnlyCache) Hit() int {
	c.t.Helper()
	return c.hit
}
