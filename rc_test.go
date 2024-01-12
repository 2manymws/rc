package rc_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/2manymws/rc"
	"github.com/2manymws/rc/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRC(t *testing.T) {
	tests := []struct {
		name          string
		cacher        rc.Cacher
		requests      []*http.Request
		wantResponses []*http.Response
		wantHit       int
	}{
		{"all cache", testutil.NewAllCache(t), []*http.Request{
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/2")},
		}, []*http.Response{
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":1}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}, "X-Cache": []string{"HIT"}}, Body: testutil.NewBody(`{"count":1}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}, "X-Cache": []string{"HIT"}}, Body: testutil.NewBody(`{"count":1}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":2}`)},
		}, 2},
		{"all cache 2", testutil.NewAllCache(t), []*http.Request{
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/2")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/1")},
		}, []*http.Response{
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":1}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":2}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}, "X-Cache": []string{"HIT"}}, Body: testutil.NewBody(`{"count":1}`)},
		}, 1},
		{"get only", testutil.NewGetOnlyCache(t), []*http.Request{
			{Method: http.MethodPost, URL: testutil.MustParseURL("http://example.com/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/1")},
			{Method: http.MethodDelete, URL: testutil.MustParseURL("http://example.com/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/1")},
		}, []*http.Response{
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":1}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":2}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":3}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}, "X-Cache": []string{"HIT"}}, Body: testutil.NewBody(`{"count":2}`)},
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := testutil.NewHTTPRouter(t)
			m := rc.New(tt.cacher)
			ts := httptest.NewServer(m(tr))
			tu := testutil.MustParseURL(ts.URL)
			t.Cleanup(ts.Close)
			tc := ts.Client()
			for i, req := range tt.requests {
				req.URL.Scheme = tu.Scheme
				req.URL.Host = tu.Host
				got, err := tc.Do(req)
				if err != nil {
					t.Fatal(err)
					return
				}
				opts := []cmp.Option{
					cmpopts.IgnoreFields(http.Response{}, "Status", "Proto", "ProtoMajor", "ProtoMinor", "ContentLength", "TransferEncoding", "Uncompressed", "Trailer", "Request", "Close", "Body"),
				}
				// header ignore fields
				got.Header.Del("Content-Length")
				got.Header.Del("Date")

				if diff := cmp.Diff(tt.wantResponses[i], got, opts...); diff != "" {
					t.Error(diff)
				}
				gotb, err := io.ReadAll(got.Body)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					got.Body.Close()
				})
				wantb, err := io.ReadAll(tt.wantResponses[i].Body)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					tt.wantResponses[i].Body.Close()
				})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(string(wantb), string(gotb)); diff != "" {
					t.Error(diff)
				}
			}

			got := tt.cacher.(testutil.Cacher).Hit()
			if got != tt.wantHit {
				t.Errorf("got %v want %v", got, tt.wantHit)
			}
		})
	}
}

func TestWithReverseProxy(t *testing.T) {
	tests := []struct {
		name            string
		useReverseProxy bool
		requests        []*http.Request
		wantResponses   []*http.Response
		wantHit         int
	}{
		{"without reverse proxy", false, []*http.Request{
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/path/to/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/path/to/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/path/to/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/path/to/2")},
		}, []*http.Response{
			{StatusCode: http.StatusNotFound, Header: http.Header{}, Body: testutil.NewBody(``)},
			{StatusCode: http.StatusNotFound, Header: http.Header{}, Body: testutil.NewBody(``)},
			{StatusCode: http.StatusNotFound, Header: http.Header{}, Body: testutil.NewBody(``)},
			{StatusCode: http.StatusNotFound, Header: http.Header{}, Body: testutil.NewBody(``)},
		}, 0},
		{"cache with reverse proxy", true, []*http.Request{
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/path/to/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/path/to/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/path/to/1")},
			{Method: http.MethodGet, URL: testutil.MustParseURL("http://example.com/path/to/2")},
		}, []*http.Response{
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":1}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}, "X-Cache": []string{"HIT"}}, Body: testutil.NewBody(`{"count":1}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}, "X-Cache": []string{"HIT"}}, Body: testutil.NewBody(`{"count":1}`)},
			{StatusCode: http.StatusOK, Header: http.Header{"Cache-Control": []string{"max-age=60"}, "Content-Type": []string{"application/json"}}, Body: testutil.NewBody(`{"count":2}`)},
		}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := testutil.NewHTTPRouter(t)
			var handler http.Handler
			if tt.useReverseProxy {
				p := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						r.URL.Path = strings.TrimPrefix(r.URL.Path, "/path/to")
						next.ServeHTTP(w, r)
					})
				}
				handler = p(tr)
			} else {
				handler = tr
			}
			cacher := testutil.NewAllCache(t)
			m := rc.New(cacher)
			ts := httptest.NewServer(m(handler))
			tu := testutil.MustParseURL(ts.URL)
			t.Cleanup(ts.Close)
			tc := ts.Client()
			for i, req := range tt.requests {
				req.URL.Scheme = tu.Scheme
				req.URL.Host = tu.Host
				got, err := tc.Do(req)
				if err != nil {
					t.Fatal(err)
					return
				}
				opts := []cmp.Option{
					cmpopts.IgnoreFields(http.Response{}, "Status", "Proto", "ProtoMajor", "ProtoMinor", "ContentLength", "TransferEncoding", "Uncompressed", "Trailer", "Request", "Close", "Body"),
				}
				// header ignore fields
				got.Header.Del("Content-Length")
				got.Header.Del("Date")

				if diff := cmp.Diff(tt.wantResponses[i], got, opts...); diff != "" {
					t.Error(diff)
				}
				gotb, err := io.ReadAll(got.Body)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					got.Body.Close()
				})
				wantb, err := io.ReadAll(tt.wantResponses[i].Body)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					tt.wantResponses[i].Body.Close()
				})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(string(wantb), string(gotb)); diff != "" {
					t.Error(diff)
				}
			}

			got := cacher.Hit()
			if got != tt.wantHit {
				t.Errorf("got %v want %v", got, tt.wantHit)
			}
		})
	}
}
