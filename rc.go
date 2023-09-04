package rc

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
)

type Cacher interface {
	// Name returns the name of the cacher.
	Name() string
	// Load loads the response cache.
	// If the cache is not found, it returns ErrCacheNotFound.
	// If not caching, it returns ErrNoCache.
	// If the cache is expired, it returns ErrCacheExpired.
	Load(req *http.Request) (res *http.Response, err error)
	// Store stores the response cache.
	Store(req *http.Request, res *http.Response) error
}

type cacheMw struct {
	cachers []Cacher
}

func newCacheMw(cachers []Cacher) *cacheMw {
	return &cacheMw{
		cachers: cachers,
	}
}

func (rc *cacheMw) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			res  *http.Response
			body []byte
			err  error
		)
		// Use cache
		hit := ""
		for _, c := range rc.cachers {
			res, err = c.Load(r)
			if err != nil {
				// TODO: log error
				continue
			}
			// Response cache
			for k, v := range res.Header {
				setted := false
				for _, vv := range v {
					if !setted {
						w.Header().Set(k, vv)
						setted = true
						continue
					}
					w.Header().Add(k, vv)
				}
			}
			w.WriteHeader(res.StatusCode)
			body, err = io.ReadAll(res.Body)
			if err == nil {
				_ = res.Body.Close()
				_, _ = w.Write(body)
			}
			_ = res.Body.Close()
			hit = c.Name()
			break
		}

		if hit == "" {
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)
			for k, v := range rec.Header() {
				setted := false
				for _, vv := range v {
					if !setted {
						w.Header().Set(k, vv)
						setted = true
						continue
					}
					w.Header().Add(k, vv)
				}
			}
			w.WriteHeader(rec.Code)
			_, _ = w.Write(rec.Body.Bytes())
			body = rec.Body.Bytes()
			res = &http.Response{
				StatusCode: rec.Code,
				Header:     rec.Header(),
			}
		}

		// Store cache
		for _, c := range rc.cachers {
			if c.Name() == hit {
				break
			}
			// Restore body
			res.Body = io.NopCloser(bytes.NewReader(body))
			// TODO: log error
			_ = c.Store(r, res)
		}
	})
}

// New returns a new response cache middleware.
// The order of the cachers is arranged in order of high-speed cache, such as CPU L1 cache and L2 cache.
func New(cachers ...Cacher) func(next http.Handler) http.Handler {
	rl := newCacheMw(cachers)
	return rl.Handler
}
