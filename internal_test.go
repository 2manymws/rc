package rc

import (
	"io"
	"net/http"
	"net/http/httptest"
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
