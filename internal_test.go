package rc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDuplicateRequest(t *testing.T) {
	tests := []struct {
		name           string
		useRequestBody bool
		wantBody       string
	}{
		{"without request body", false, ""},
		{"with request body", true, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &cacheMw{
				useRequestBody: tt.useRequestBody,
			}
			req := httptest.NewRequest(http.MethodGet, "http://example.com/path/to/1", strings.NewReader("hello"))
			req, preq := m.duplicateRequest(req)
			b, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			if err := req.Body.Close(); err != nil {
				t.Fatal(err)
			}
			if string(b) != "hello" {
				t.Errorf("got %s want %s", b, "hello")
			}
			{
				b, err := io.ReadAll(preq.Body)
				if err != nil {
					t.Fatal(err)
				}
				if err := preq.Body.Close(); err != nil {
					t.Fatal(err)
				}
				if string(b) != tt.wantBody {
					t.Errorf("got %s want %s", b, "hello")
				}
			}
		})
	}
}

func TestMaskHeader(t *testing.T) {
	tests := []struct {
		h                 http.Header
		headerNamesToMask []string
		want              http.Header
	}{
		{nil, nil, nil},
		{
			http.Header{
				"Date": []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
			},
			defaultHeaderNamesToMask,
			http.Header{
				"Date": []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
			},
		},
		{
			http.Header{
				"Date":       []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie": []string{"session=1234; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; HttpOnly"},
			},
			defaultHeaderNamesToMask,
			http.Header{
				"Date":       []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie": []string{"*****"},
			},
		},
		{
			http.Header{
				"Date": []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie": []string{
					"session=1234; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; HttpOnly",
					"session=5678; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; HttpOnly",
				},
			},
			defaultHeaderNamesToMask,
			http.Header{
				"Date":       []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie": []string{"*****"},
			},
		},
		{
			http.Header{
				"Date":          []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie":    []string{"session=1234; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; HttpOnly"},
				"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9."},
			},
			defaultHeaderNamesToMask,
			http.Header{
				"Date":          []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie":    []string{"*****"},
				"Authorization": []string{"*****"},
			},
		},
		{
			http.Header{
				"Date":       []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie": []string{"session=1234; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; HttpOnly"},
				"Authorization": []string{
					"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.",
					"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.",
				},
			},
			defaultHeaderNamesToMask,
			http.Header{
				"Date":          []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie":    []string{"*****"},
				"Authorization": []string{"*****"},
			},
		},
		{
			http.Header{
				"Date":          []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie":    []string{"session=1234; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; HttpOnly"},
				"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9."},
			},
			nil,
			http.Header{
				"Date":          []string{"Mon, 02 Jan 2006 15:04:05 GMT"},
				"Set-Cookie":    []string{"session=1234; Path=/; Expires=Wed, 09 Jun 2021 10:18:14 GMT; HttpOnly"},
				"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9."},
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			m := &cacheMw{
				headerNamesToMask: defaultHeaderNamesToMask,
			}
			HeaderNamesToMask(tt.headerNamesToMask)(m)
			got := m.maskHeader(tt.h)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}
