package testutil

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

var (
	_ Cacher = &AllCache{}
	_ Cacher = &GetOnlyCache{}
)

// errCacheNotFound is returned when the cache is not found
var errCacheNotFound error = errors.New("cache not found")

// errNoCache is returned if not caching
var errNoCache error = errors.New("no cache")

// errCacheExpired is returned if the cache is expired
var errCacheExpired error = errors.New("cache expired")

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
}

type GetOnlyCache struct {
	t   testing.TB
	m   map[string]string
	dir string
	hit int
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
	p, ok := c.m[req.URL.String()]
	if !ok {
		return nil, errCacheNotFound
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, errCacheNotFound
	}
	res, err = bytesToResponse(b)
	if err != nil {
		return nil, err
	}
	c.hit++
	return res, nil
}

func (c *AllCache) Store(req *http.Request, res *http.Response) error {
	c.t.Helper()
	b, err := responseToBytes(res)
	if err != nil {
		return err
	}
	p := filepath.Join(c.dir, base64.URLEncoding.EncodeToString([]byte(req.URL.String())))
	if err := os.WriteFile(p, b, os.ModePerm); err != nil {
		return err
	}
	c.m[req.URL.String()] = p
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
	p, ok := c.m[req.URL.String()]
	if !ok {
		return nil, errCacheNotFound
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, errCacheNotFound
	}
	res, err = bytesToResponse(b)
	if err != nil {
		return nil, err
	}
	c.hit++
	return res, nil
}

func (c *GetOnlyCache) Store(req *http.Request, res *http.Response) error {
	c.t.Helper()
	if req.Method != http.MethodGet {
		return errNoCache
	}
	b, err := responseToBytes(res)
	if err != nil {
		return err
	}
	p := filepath.Join(c.dir, base64.URLEncoding.EncodeToString([]byte(req.URL.String())))
	if err := os.WriteFile(p, b, os.ModePerm); err != nil {
		return err
	}
	c.m[req.URL.String()] = p
	return nil
}

func (c *GetOnlyCache) Hit() int {
	c.t.Helper()
	return c.hit
}

type cacheResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func responseToBytes(res *http.Response) ([]byte, error) {
	c := &cacheResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	c.Body = b
	cb, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return cb, nil
}

func bytesToResponse(b []byte) (*http.Response, error) {
	c := &cacheResponse{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	res := &http.Response{
		StatusCode: c.StatusCode,
		Header:     c.Header,
		Body:       io.NopCloser(bytes.NewReader(c.Body)),
	}
	return res, nil
}
