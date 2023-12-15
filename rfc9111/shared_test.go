package rfc9111

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestShared_Storable(t *testing.T) {
	now := time.Date(2024, 12, 13, 14, 15, 16, 00, time.UTC)

	tests := []struct {
		name        string
		req         *http.Request
		res         *http.Response
		wantOK      bool
		wantExpires time.Time
	}{
		{
			"GET 200 Cache-Control: s-maxage=10 -> +10s",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Cache-Control": []string{"s-maxage=10"},
				},
			},
			true,
			time.Date(2024, 12, 13, 14, 15, 26, 00, time.UTC),
		},
		{
			"GET 200 Cache-Control: max-age=15 -> +15s",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Cache-Control": []string{"max-age=15"},
				},
			},
			true,
			time.Date(2024, 12, 13, 14, 15, 31, 00, time.UTC),
		},
		{
			"GET 200 Expires: 2024-12-13 14:15:20",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Expires": []string{"Mon, 13 Dec 2024 14:15:20 GMT"},
				},
			},
			true,
			time.Date(2024, 12, 13, 14, 15, 20, 00, time.UTC),
		},
		{
			"GET 200 Expires: 2024-12-13 14:15:20, Date: 2024-12-13 13:15:20",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Expires": []string{"Mon, 13 Dec 2024 14:15:20 GMT"},
					"Date":    []string{"Mon, 13 Dec 2024 13:15:20 GMT"},
				},
			},
			true,
			time.Date(2024, 12, 13, 15, 15, 16, 00, time.UTC),
		},
		{
			"GET 200 Last-Modified: 2024-12-13 14:15:10, Date: 2024-12-13 14:15:20 -> +1s",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{"Mon, 13 Dec 2024 14:15:10 GMT"},
					"Date":          []string{"Mon, 13 Dec 2024 14:15:20 GMT"},
				},
			},
			true,
			time.Date(2024, 12, 13, 14, 15, 21, 00, time.UTC),
		},
		{
			"GET 200 Last-Modified: 2024-12-13 14:15:06 -> +1s",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{"Mon, 13 Dec 2024 14:15:06 GMT"},
				},
			},
			true,
			time.Date(2024, 12, 13, 14, 15, 17, 00, time.UTC),
		},
		{
			"GET 500 Last-Modified: 2024-12-13 14:15:06 -> No Store",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusInternalServerError,
				Header: http.Header{
					"Last-Modified": []string{"Mon, 13 Dec 2024 14:15:06 GMT"},
				},
			},
			false,
			time.Time{},
		},
		{
			"GET 200 -> No Store",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Date": []string{"Mon, 13 Dec 2024 14:15:10 GMT"},
				},
			},
			false,
			time.Time{},
		},
		{
			"UNUNDERSTOODMETHOD 200 -> No Store",
			&http.Request{
				Method: "UNUNDERSTOODMETHOD",
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Cache-Control": []string{"max-age=15"},
				},
			},
			false,
			time.Time{},
		},
		{
			"GET 100 -> No Store",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusContinue,
				Header: http.Header{
					"Cache-Control": []string{"max-age=15"},
				},
			},
			false,
			time.Time{},
		},
		{
			"GET 206 -> No Store",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusPartialContent,
				Header: http.Header{
					"Cache-Control": []string{"max-age=15"},
				},
			},
			false,
			time.Time{},
		},
		{
			"GET 200 Cache-Control: no-store -> No Store",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Cache-Control": []string{"no-store"},
				},
			},
			false,
			time.Time{},
		},
		{
			"GET 200 Cache-Control: private -> No Store",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Cache-Control": []string{"private"},
				},
			},
			false,
			time.Time{},
		},
		{
			"GET 200 Cache-Control: public, Last-Modified 2024-12-13 14:15:06 -> +1s",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{"Mon, 13 Dec 2024 14:15:06 GMT"},
					"Cache-Control": []string{"public"},
				},
			},
			true,
			time.Date(2024, 12, 13, 14, 15, 17, 00, time.UTC),
		},
		{
			"POST 201 Cache-Control: public, Last-Modified 2024-12-13 14:15:06 -> No Store",
			&http.Request{
				Method: http.MethodPost,
			},
			&http.Response{
				StatusCode: http.StatusCreated,
				Header: http.Header{
					"Last-Modified": []string{"Mon, 13 Dec 2024 14:15:06 GMT"},
					"Cache-Control": []string{"public"},
				},
			},
			false,
			time.Time{},
		},
		{
			"GET 200 Cache-Control: public (only) -> No Store",
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Cache-Control": []string{"public"},
				},
			},
			false,
			time.Time{},
		},
		{
			"GET Authorization: XXX 200 Cache-Control: max-age=15 -> No Store",
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Authorization": []string{"XXX"},
				},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Cache-Control": []string{"max-age=15"},
				},
			},
			false,
			time.Time{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := NewShared()
			if err != nil {
				t.Errorf("Shared.Storable() error = %v", err)
				return
			}
			gotOK, gotExpires := s.Storable(tt.req, tt.res, now)
			if gotOK != tt.wantOK {
				t.Errorf("Shared.Storable() gotOK = %v, want %v", gotOK, tt.wantOK)
			}
			if !gotExpires.Equal(tt.wantExpires) {
				t.Errorf("Shared.Storable() gotExpires = %v, want %v", gotExpires, tt.wantExpires)
			}
		})
	}
}

func TestShared_Handle(t *testing.T) {
	// 2024-12-13 14:15:16
	now := time.Date(2024, 12, 13, 14, 15, 16, 00, time.UTC)
	before15sec := now.Add(-15 * time.Second)
	before30sec := now.Add(-30 * time.Second)

	endpoint, err := url.Parse("https://example.com/api/v1/path/to/resource")
	if err != nil {
		t.Fatal(err)
	}
	other, err := url.Parse("https://example.com/api/v2/path/to/resource")
	if err != nil {
		t.Fatal(err)
	}

	origin200res := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Age": []string{"0"},
		},
	}
	do200 := func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
		}, nil
	}
	do304 := func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotModified,
			Header:     http.Header{},
		}, nil
	}
	tests := []struct {
		name          string
		req           *http.Request
		cachedReq     *http.Request
		cachedRes     *http.Response
		do            func(req *http.Request) (*http.Response, error)
		wantCacheUsed bool
		wantRes       *http.Response
	}{
		{
			"No cached request/response (nil)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			nil,
			nil,
			do200,
			false,
			origin200res,
		},
		{
			"Use origin response (defferent URL)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Request{
				URL:    other,
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
				},
			},
			do200,
			false,
			origin200res,
		},
		{
			"Use origin response (defferent method)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodHead,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
				},
			},
			do200,
			false,
			origin200res,
		},
		{
			"Use origin response (Vary: *)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Vary":          []string{"*"},
					"Cache-Control": []string{"max-age=20"},
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
				},
			},
			do200,
			false,
			origin200res,
		},
		{
			"Use flesh cached response",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
					"Date":          []string{before15sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"max-age=20"},
				},
			},
			do200,
			true,
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Age":           []string{"15"},
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
					"Date":          []string{before15sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"max-age=20"},
				},
			},
		},
		{
			"Use stale cached response",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Date":          []string{before30sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"max-age=20"},
				},
			},
			do200,
			true,
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Age":           []string{"30"},
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Date":          []string{before30sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"max-age=20"},
				},
			},
		},
		{
			"Use origin response (Cache-Control: max-stale=10)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{
					"Cache-Control": []string{"max-stale=10"},
				},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Date":          []string{before30sec.Format(http.TimeFormat)},
				},
			},
			do200,
			false,
			origin200res,
		},
		{
			"Use stale cached response (Cache-Control: max-stale=40)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{
					"Cache-Control": []string{"max-stale=40"},
				},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Date":          []string{before30sec.Format(http.TimeFormat)},
				},
			},
			do200,
			true,
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Age":           []string{"30"},
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Date":          []string{before30sec.Format(http.TimeFormat)},
				},
			},
		},
		{
			"Use flesh cached response (Vary: Content-Type, User-Agent)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{
					"User-Agent":   []string{"test"},
					"Content-Type": []string{"application/json"},
				},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
					"User-Agent":   []string{"test"},
				},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
					"Date":          []string{before15sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"max-age=20"},
					"Vary":          []string{"content-type, user-agent"},
				},
			},
			do200,
			true,
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Age":           []string{"15"},
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
					"Date":          []string{before15sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"max-age=20"},
					"Vary":          []string{"content-type, user-agent"},
				},
			},
		},
		{
			"Use origin response (Vary: Content-Type, User-Agent)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{
					"User-Agent":   []string{"test2"},
					"Content-Type": []string{"application/json"},
				},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
					"User-Agent":   []string{"test"},
				},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"max-age=20"},
					"Vary":          []string{"content-type, user-agent"},
				},
			},
			do200,
			false,
			origin200res,
		},
		{
			"Use origin response (Cache-Control: no-cache, max-age=60)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"no-cache, max-age=60"},
				},
			},
			do200,
			false,
			origin200res,
		},
		{
			"Use origin response (POST Cache-Control: no-cache)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodPost,
				Header: http.Header{},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodPost,
				Header: http.Header{},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"no-cache"},
				},
			},
			do200,
			false,
			origin200res,
		},
		{
			"Use cache response (Cache-Control: no-cache, max-age=60)",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"no-cache, max-age=60"},
				},
			},
			do304,
			true,
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before15sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"no-cache, max-age=60"},
				},
			},
		},
		{
			"Validate and use origin response",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Date":          []string{before30sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"must-revalidate"},
				},
			},
			do200,
			false,
			origin200res,
		},
		{
			"Validate and use cached response",
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{},
			},
			&http.Request{
				URL:    endpoint,
				Method: http.MethodGet,
				Header: http.Header{},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Date":          []string{before30sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"must-revalidate"},
				},
			},
			do304,
			true,
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Age":           []string{"30"},
					"Last-Modified": []string{before30sec.Format(http.TimeFormat)},
					"Date":          []string{before30sec.Format(http.TimeFormat)},
					"Cache-Control": []string{"must-revalidate"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			//t.Parallel()
			s, err := NewShared()
			if err != nil {
				t.Errorf("Shared.Handle() error = %v", err)
				return
			}
			gotCacheUsed, gotRes, err := s.Handle(tt.req, tt.cachedReq, tt.cachedRes, tt.do, now)
			if err != nil {
				t.Errorf("Shared.Handle() error = %v", err)
				return
			}
			if gotCacheUsed != tt.wantCacheUsed {
				t.Errorf("Shared.Handle() gotCacheUsed = %v, want %v", gotCacheUsed, tt.wantCacheUsed)
			}
			if diff := cmp.Diff(gotRes, tt.wantRes); diff != "" {
				t.Errorf("Shared.Handle() gotRes != tt.wantRes:\n%s", diff)
			}
		})
	}
}

func TestShared_SharedOption(t *testing.T) {
	now := time.Date(2024, 12, 13, 14, 15, 16, 00, time.UTC)

	tests := []struct {
		name        string
		opts        []SharedOption
		req         *http.Request
		res         *http.Response
		wantOK      bool
		wantExpires time.Time
	}{
		{
			"GET 200 Last-Modified: 2024-12-13 14:15:06 -> +1s",
			[]SharedOption{},
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{"Mon, 13 Dec 2024 14:15:06 GMT"},
				},
			},
			true,
			time.Date(2024, 12, 13, 14, 15, 17, 00, time.UTC),
		},
		{
			"HeuristicExpirationRatio 0.5 GET 200 Last-Modified: 2024-12-13 14:15:06 -> +5s",
			[]SharedOption{
				HeuristicExpirationRatio(0.5),
			},
			&http.Request{
				Method: http.MethodGet,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Last-Modified": []string{"Mon, 13 Dec 2024 14:15:06 GMT"},
				},
			},
			true,
			time.Date(2024, 12, 13, 14, 15, 21, 00, time.UTC),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := NewShared(tt.opts...)
			if err != nil {
				t.Errorf("Shared.Storable() error = %v", err)
				return
			}
			gotOK, gotExpires := s.Storable(tt.req, tt.res, now)
			if gotOK != tt.wantOK {
				t.Errorf("Shared.Storable() gotOK = %v, want %v", gotOK, tt.wantOK)
			}
			if !gotExpires.Equal(tt.wantExpires) {
				t.Errorf("Shared.Storable() gotExpires = %v, want %v", gotExpires, tt.wantExpires)
			}
		})
	}
}
