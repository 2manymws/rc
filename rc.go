package rc

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/2manymws/rc/rfc9111"
)

var defaultHeaderNamesToMask = []string{
	"Authorization",
	"Cookie",
	"Set-Cookie",
}

type Cacher interface {
	// Load loads the request/response cache.
	// If the cache is not found, it returns ErrCacheNotFound.
	// If the cache is expired, it returns ErrCacheExpired.
	// If not caching, it returns ErrShouldNotUseCache.
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
	cacher            *cacher
	useRequestBody    bool
	logger            *slog.Logger
	headerNamesToMask []string
}

func newCacheMw(c Cacher, opts ...Option) *cacheMw {
	cc := newCacher(c)
	m := &cacheMw{
		cacher:            cc,
		headerNamesToMask: defaultHeaderNamesToMask,
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
		// rc does not support websocket.
		if strings.ToLower(req.Header.Get("Connection")) == "upgrade" && strings.ToLower(req.Header.Get("Upgrade")) == "websocket" {
			next.ServeHTTP(w, req)
			return
		}
		now := time.Now()

		// Copy the request so that it is not affected by the next handler.
		// reqc is the request to be used for caching.
		req, reqc := m.duplicateRequest(req)

		cachedReq, cachedRes, err := m.cacher.Load(reqc) //nostyle:handlerrors
		if err != nil {
			switch {
			case errors.Is(err, ErrCacheNotFound):
				m.logger.Debug("cache not found", slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)))
			case errors.Is(err, ErrCacheExpired):
				m.logger.Debug("cache expired", slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)))
			case errors.Is(err, ErrShouldNotUseCache):
				m.logger.Debug("should not use cache", slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)))
				// Skip caching
				next.ServeHTTP(w, req)
				return
			default:
				m.logger.Error("failed to load cache", slog.String("error", err.Error()), slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)))
			}
		} else {
			defer func() {
				cachedReq.Body.Close()
				cachedRes.Body.Close()
			}()
		}
		cacheUsed, res, err := m.cacher.Handle(req, cachedReq, cachedRes, m.handlerToRequester(next, reqc, now), now) //nostyle:handlerrors
		if err != nil {
			m.logger.Error("failed to handle cache", slog.String("error", err.Error()), slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)))
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				m.logger.Error("failed to close response body", slog.String("error", err.Error()), slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", m.maskHeader(res.Header)))
			}
		}()

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

		ww, ok := w.(io.Writer)
		if !ok {
			m.logger.Error("failed to cast response writer to io.Writer", slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()))
			return
		}
		buf := getCopyBuf()
		defer putCopyBuf(buf)
		if _, err := io.CopyBuffer(ww, res.Body, buf); err != nil {
			// Error as debug
			// - os.ErrDeadlineExceeded: The request context has been canceled or has expired.
			// - "client disconnected": The client disconnected. (net/http.http2errClientDisconnected)
			// - "http2: stream closed": The client disconnected. (net/http.http2errStreamClosed)
			// - syscall.ECONNRESET: The client disconnected. ("connection reset by peer")
			// - syscall.EPIPE: The client disconnected. ("broken pipe")
			// - http.ErrBodyNotAllowed: The request method does not allow a body.
			switch {
			case errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) || contains([]string{"client disconnected", "http2: stream closed"}, err.Error()):
				m.logger.Debug("failed to write response body", slog.String("error", err.Error()), slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", m.maskHeader(res.Header)))
			case errors.Is(err, http.ErrBodyNotAllowed):
				// It is desirable that there should be no content body in the response, but the proxy server cannot handle it, so it is used as a debug log.
				m.logger.Debug("failed to write response body", slog.String("error", err.Error()), slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", m.maskHeader(res.Header)))
			default:
				m.logger.Error("failed to write response body", slog.String("error", err.Error()), slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", m.maskHeader(res.Header)))
			}
		}
		if cacheUsed {
			m.logger.Debug("cache used", slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)), slog.Int("status", res.StatusCode))
		}
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
		m.logger.Error("failed to read request body", slog.String("error", err.Error()), slog.String("host", copy.Host), slog.String("method", copy.Method), slog.Any("headers", m.maskHeader(copy.Header)), slog.String("url", copy.URL.String()))
	}
	if err := copy.Body.Close(); err != nil {
		m.logger.Error("failed to close request body", slog.String("error", err.Error()), slog.String("host", copy.Host), slog.String("method", copy.Method), slog.Any("headers", m.maskHeader(copy.Header)), slog.String("url", copy.URL.String()))
	}
	req.Body = io.NopCloser(bytes.NewReader(b))
	copy.Body = io.NopCloser(bytes.NewReader(b))
	return copy, req
}

func (m *cacheMw) handlerToRequester(h http.Handler, reqc *http.Request, now time.Time) func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		rec := newRecorder()
		defer rec.Reset()
		h.ServeHTTP(rec, req)
		res := rec.Result()
		resc := rec.Result()

		go func() {
			ok, expires := m.cacher.Storable(reqc, resc, now)
			if !ok {
				m.logger.Debug("cache not storable", slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)), slog.Int("status", res.StatusCode), slog.Any("response_headers", m.maskHeader(resc.Header)))
				return
			}

			// Store response as cache
			if err := m.cacher.Store(reqc, resc, expires); err != nil {
				m.logger.Error("failed to store cache", slog.String("error", err.Error()), slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)), slog.Int("status", resc.StatusCode))
			}
			m.logger.Debug("cache stored", slog.String("host", reqc.Host), slog.String("method", reqc.Method), slog.String("url", reqc.URL.String()), slog.Any("headers", m.maskHeader(reqc.Header)), slog.Int("status", resc.StatusCode))
		}()

		return res, nil
	}
}

func (m *cacheMw) maskHeader(h http.Header) http.Header {
	const masked = "*****"
	c := h.Clone()
	for _, n := range m.headerNamesToMask {
		if c.Get(n) != "" {
			c.Set(n, masked)
		}
	}
	return c
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

// HeaderNamesToMask sets header names to mask in logs.
func HeaderNamesToMask(names []string) Option {
	return func(m *cacheMw) {
		m.headerNamesToMask = names
	}
}

// New returns a new response cache middleware.
func New(cacher Cacher, opts ...Option) func(next http.Handler) http.Handler {
	rl := newCacheMw(cacher, opts...)
	return rl.Handler
}

type recorder struct {
	statusCode int
	header     http.Header
	buf        *bytes.Buffer
}

var _ http.ResponseWriter = (*recorder)(nil)

func newRecorder() *recorder {
	return &recorder{
		buf:    new(bytes.Buffer),
		header: make(http.Header),
	}
}

func (r *recorder) Header() http.Header {
	return r.header
}

func (r *recorder) Write(b []byte) (int, error) {
	return r.buf.Write(b)
}

func (r *recorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *recorder) Result() *http.Response {
	return &http.Response{
		Status:        http.StatusText(r.statusCode),
		StatusCode:    r.statusCode,
		Header:        r.header.Clone(),
		Body:          io.NopCloser(bytes.NewReader(r.buf.Bytes())),
		ContentLength: int64(r.buf.Len()),
	}
}

func (r *recorder) Reset() {
	r.statusCode = 0
	r.header = make(http.Header)
	r.buf.Reset()
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}
