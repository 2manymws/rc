package testutil

import (
	"net/http"

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
	m   map[string][]byte
	hit int
}

type GetOnlyCache struct {
	m   map[string][]byte
	hit int
}

func NewAllCache() *AllCache {
	return &AllCache{
		m: map[string][]byte{},
	}
}

func (c *AllCache) Name() string {
	return "all"
}

func (c *AllCache) Load(req *http.Request) (res *http.Response, err error) {
	b, ok := c.m[req.URL.String()]
	if !ok {
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
	b, err := rc.ResponseToBytes(res)
	if err != nil {
		return err
	}
	c.m[req.URL.String()] = b
	return nil
}

func (c *AllCache) Hit() int {
	return c.hit
}

func NewGetOnlyCache() *GetOnlyCache {
	return &GetOnlyCache{
		m: map[string][]byte{},
	}
}

func (c *GetOnlyCache) Name() string {
	return "get-only"
}

func (c *GetOnlyCache) Load(req *http.Request) (res *http.Response, err error) {
	if req.Method != http.MethodGet {
		return nil, rc.ErrNoCache
	}
	b, ok := c.m[req.URL.String()]
	if !ok {
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
	if req.Method != http.MethodGet {
		return rc.ErrNoCache
	}
	b, err := rc.ResponseToBytes(res)
	if err != nil {
		return err
	}
	c.m[req.URL.String()] = b
	return nil
}

func (c *GetOnlyCache) Hit() int {
	return c.hit
}
