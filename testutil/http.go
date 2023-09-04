package testutil

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/k1LoW/httpstub"
)

// NewTestHTTPRouter returns a new test HTTP router
func NewTestHTTPRouter(t *testing.T) *httpstub.Router {
	r := httpstub.NewRouter(t)
	count := 0
	r.Path("/no-store-header").Handler(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"count":%d}`, count)))
	})
	r.Match(func(r *http.Request) bool { return true }).Handler(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"count":%d}`, count)))
	})
	return r
}

func MustParseURL(urlstr string) *url.URL {
	u, err := url.Parse(urlstr)
	if err != nil {
		panic(err)
	}
	return u
}

func NewBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func ReadBody(r io.ReadCloser) string {
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}
