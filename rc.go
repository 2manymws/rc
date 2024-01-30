package rc

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/2manymws/rc/rfc9111"
)

type Cacher interface {
	// Load loads the request/response cache.
	// If the cache is not found, it returns ErrCacheNotFound.
	// If not caching, it returns ErrNoCache.
	// If the cache is expired, it returns ErrCacheExpired.
	Load(req *http.Request) (cachedReq *http.Request, cachedRes *http.Response, err error)
	// Store stores the response cache.
	Store(req *http.Request, res *http.Response, expires time.Time) error
}

type Handler interface {
	// Handle handles the request/response cache.
	Handle(req *http.Request, cachedReq *http.Request, cachedRes *http.Response, originRequester func(*http.Request) (*http.Response, error), now time.Time) (cacheUsed bool, res *http.Response, err error)
	// Storable returns whether the response is storable and the expiration time.
	Storable(req *http.Request, res *http.Response, now time.Time) (ok bool, expires time.Time)
}

var _ Handler = (*rfc9111.Shared)(nil)

type cacher struct {
	Cacher
	Handle   func(req *http.Request, cachedReq *http.Request, cachedRes *http.Response, originRequester func(*http.Request) (*http.Response, error), now time.Time) (cacheUsed bool, res *http.Response, err error)
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
		s, err := rfc9111.NewShared()
		if err != nil {
			panic(err) //nostyle:dontpanic
		}
		cc.Handle = s.Handle
		cc.Storable = s.Storable
	}
	return cc
}

type cacheMw struct {
	cacher         *cacher
	useRequestBody bool
	logger         *slog.Logger
}

func newCacheMw(c Cacher, opts ...Option) *cacheMw {
	cc := newCacher(c)
	m := &cacheMw{
		cacher: cc,
	}
	for _, opt := range opts {
		opt(m)
	}
	if m.logger == nil {
		m.logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	return m
}

func (m *cacheMw) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		now := time.Now()

		// Copy the request so that it is not affected by the next handler.
		req, preq := m.duplicateRequest(req)

		cachedReq, cachedRes, err := m.cacher.Load(preq) //nostyle:handlerrors
		if err != nil {
			switch {
			case errors.Is(err, ErrCacheNotFound):
				m.logger.Debug("cache not found", slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)))
			case errors.Is(err, ErrCacheExpired):
				m.logger.Warn("cache expired", slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)))
			case errors.Is(err, ErrShouldNotUseCache):
				m.logger.Debug("should not use cache", slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)))
			default:
				m.logger.Error("failed to load cache", slog.String("error", err.Error()), slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)))
			}
		}
		cacheUsed, res, err := m.cacher.Handle(req, cachedReq, cachedRes, HandlerToRequester(next), now) //nostyle:handlerrors
		if err != nil {
			m.logger.Error("failed to handle cache", slog.String("error", err.Error()), slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)))
		}

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
		if err != nil {
			m.logger.Error("failed to read response body", slog.String("error", err.Error()), slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", maskHeader(res.Header)))
		} else {
			if _, err := w.Write(body); err != nil {
				if errors.Is(err, http.ErrBodyNotAllowed) {
					m.logger.Warn("failed to write response body", slog.String("error", err.Error()), slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", maskHeader(res.Header)))
				} else {
					m.logger.Error("failed to write response body", slog.String("error", err.Error()), slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", maskHeader(res.Header)))
				}
			}
		}
		if err := res.Body.Close(); err != nil {
			m.logger.Error("failed to close response body", slog.String("error", err.Error()), slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", maskHeader(res.Header)))
		}

		if cacheUsed {
			m.logger.Debug("cache used", slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)), slog.Int("status", res.StatusCode))
			return
		}
		ok, expires := m.cacher.Storable(preq, res, now)
		if !ok {
			m.logger.Debug("cache not storable", slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", maskHeader(res.Header)))
			return
		}
		// Restore response body
		res.Body = io.NopCloser(bytes.NewReader(body))

		// Store response as cache
		if err := m.cacher.Store(preq, res, expires); err != nil {
			m.logger.Error("failed to store cache", slog.String("error", err.Error()), slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)), slog.Int("status", res.StatusCode))
		}
		m.logger.Debug("cache stored", slog.String("host", preq.Host), slog.String("method", preq.Method), slog.String("url", preq.URL.String()), slog.Any("headers", maskHeader(preq.Header)), slog.Int("status", res.StatusCode))
	})
}

func (m *cacheMw) duplicateRequest(req *http.Request) (*http.Request, *http.Request) {
	copy := req.Clone(req.Context())
	if !m.useRequestBody {
		// request Body is not copied since it is not used.
		// req.Body is already closed.
		return copy, req
	}
	b, err := io.ReadAll(copy.Body)
	if err != nil {
		m.logger.Error("failed to read request body", slog.String("error", err.Error()), slog.String("host", copy.Host), slog.String("method", copy.Method), slog.Any("headers", maskHeader(copy.Header)), slog.String("url", copy.URL.String()))
	}
	if err := copy.Body.Close(); err != nil {
		m.logger.Error("failed to close request body", slog.String("error", err.Error()), slog.String("host", copy.Host), slog.String("method", copy.Method), slog.Any("headers", maskHeader(copy.Header)), slog.String("url", copy.URL.String()))
	}
	req.Body = io.NopCloser(bytes.NewReader(b))
	copy.Body = io.NopCloser(bytes.NewReader(b))
	return copy, req
}

type Option func(*cacheMw)

// WithLogger sets logger (slog.Logger).
func WithLogger(l *slog.Logger) Option {
	return func(m *cacheMw) {
		m.logger = l
	}
}

// UseRequestBody enables to use request body as cache key.
func UseRequestBody() Option {
	return func(m *cacheMw) {
		m.useRequestBody = true
	}
}

// New returns a new response cache middleware.
func New(cacher Cacher, opts ...Option) func(next http.Handler) http.Handler {
	rl := newCacheMw(cacher, opts...)
	return rl.Handler
}

// HandlerToRequester converts http.Handler to func(*http.Request) (*http.Response, error).
func HandlerToRequester(h http.Handler) func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		res := rec.Result()
		res.Header = rec.Header()
		return res, nil
	}
}

func maskHeader(h http.Header) http.Header {
	const masked = "*****"
	c := h.Clone()
	if c.Get("Set-Cookie") != "" {
		c.Set("Set-Cookie", masked)
	}
	if c.Get("Authorization") != "" {
		c.Set("Authorization", masked)
	}
	return c
}
