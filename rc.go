package rc

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/k1LoW/rc/rfc9111"
)

type Cacher interface {
	// Load loads the request/response cache.
	// If the cache is not found, it returns ErrCacheNotFound.
	// If not caching, it returns ErrNoCache.
	// If the cache is expired, it returns ErrCacheExpired.
	Load(req *http.Request) (cachedReq *http.Request, cachedRes *http.Response, err error)
	// Store stores the response cache.
	Store(req *http.Request, res *http.Response) error
}

type Handler interface {
	// Handle handles the request/response cache.
	Handle(req *http.Request, cachedReq *http.Request, cachedRes *http.Response, do func(*http.Request) (*http.Response, error), now time.Time) (cacheUsed bool, res *http.Response, err error)
	// Storable returns whether the response is storable and the expiration time.
	Storable(req *http.Request, res *http.Response, now time.Time) (ok bool, expires time.Time)
}

var _ Handler = new(rfc9111.Shared)

type cacher struct {
	Cacher
	Handle   func(req *http.Request, cachedReq *http.Request, cachedRes *http.Response, do func(*http.Request) (*http.Response, error), now time.Time) (cacheUsed bool, res *http.Response, err error)
	Storable func(req *http.Request, res *http.Response, now time.Time) (ok bool, expires time.Time)
}

func newCacher(c Cacher) *cacher {
	cc := &cacher{
		Cacher: c,
	}
	if v, ok := c.(Handler); ok {
		cc.Handle = v.Handle
		cc.Storable = v.Storable
	} else {
		s, _ := rfc9111.NewShared()
		cc.Handle = s.Handle
		cc.Storable = s.Storable
	}
	return cc
}

type cacheMw struct {
	cacher *cacher
}

func newCacheMw(c Cacher) *cacheMw {
	cc := newCacher(c)
	return &cacheMw{
		cacher: cc,
	}
}

func (cw *cacheMw) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		now := time.Now()
		cachedReq, cachedRes, _ := cw.cacher.Load(req)
		cacheUsed, res, _ := cw.cacher.Handle(req, cachedReq, cachedRes, HandlerToClientDo(next), now)

		// Response
		for k, v := range res.Header {
			set := false
			for _, vv := range v {
				if !set {
					w.Header().Set(k, vv)
					set = true
					continue
				}
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(res.StatusCode)
		body, err := io.ReadAll(res.Body)
		if err == nil {
			_ = res.Body.Close()
			_, _ = w.Write(body)
		}
		_ = res.Body.Close()

		if cacheUsed {
			return
		}
		ok, _ := cw.cacher.Storable(req, res, now)
		if !ok {
			return
		}
		// Restore response body
		res.Body = io.NopCloser(bytes.NewReader(body))

		// Store response as cache
		_ = cw.cacher.Store(req, res)
	})
}

// New returns a new response cache middleware.
func New(cacher Cacher) func(next http.Handler) http.Handler {
	rl := newCacheMw(cacher)
	return rl.Handler
}

// HandlerToClientDo converts http.Handler to func(*http.Request) (*http.Response, error).
func HandlerToClientDo(h http.Handler) func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		res := rec.Result()
		res.Header = rec.Header()
		return res, nil
	}
}
