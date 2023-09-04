package testutil

import (
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/rc"
)

var (
	_ rc.Cacher = &AllCache{}
	_ rc.Cacher = &GetOnlyCache{}

	_ Cacher = &AllCache{}
	_ Cacher = &GetOnlyCache{}
)

type Cacher interface {
	rc.Cacher
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
		return nil, rc.ErrCacheNotFound
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, rc.ErrCacheNotFound
	}
	res, err = rc.BytesToResponse(b)
	if err != nil {
		return nil, err
	}
	c.hit++
	return res, nil
}

func (c *AllCache) Store(req *http.Request, res *http.Response) error {
	c.t.Helper()
	b, err := rc.ResponseToBytes(res)
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
		return nil, rc.ErrNoCache
	}
	p, ok := c.m[req.URL.String()]
	if !ok {
		return nil, rc.ErrCacheNotFound
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, rc.ErrCacheNotFound
	}
	res, err = rc.BytesToResponse(b)
	if err != nil {
		return nil, err
	}
	c.hit++
	return res, nil
}

func (c *GetOnlyCache) Store(req *http.Request, res *http.Response) error {
	c.t.Helper()
	if req.Method != http.MethodGet {
		return rc.ErrNoCache
	}
	b, err := rc.ResponseToBytes(res)
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
