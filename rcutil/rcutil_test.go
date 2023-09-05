package rcutil

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestResponseToBytesToResponse(t *testing.T) {
	tests := []struct {
		res  *http.Response
		want *http.Response
	}{
		{
			res:  &http.Response{Body: newBody("")},
			want: &http.Response{Body: newBody("")},
		},
		{
			res:  &http.Response{Status: http.StatusText(http.StatusOK), StatusCode: http.StatusOK, Body: newBody("")},
			want: &http.Response{StatusCode: http.StatusOK, Body: newBody("")},
		},
		{
			res:  &http.Response{Header: http.Header{"X-Cache": []string{"HIT"}, "X-Hello": []string{"World"}}, Body: newBody("")},
			want: &http.Response{Header: http.Header{"X-Cache": []string{"HIT"}, "X-Hello": []string{"World"}}, Body: newBody("")},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b, err := ResponseToBytes(tt.res)
			if err != nil {
				t.Fatal(err)
			}
			got, err := BytesToResponse(b)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				got.Body.Close()
				tt.want.Body.Close()
			})
			opts := []cmp.Option{
				cmpopts.IgnoreFields(http.Response{}, "Body"),
			}
			if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
				t.Error(diff)
			}
			gotb := readBody(got.Body)
			wantb := readBody(tt.want.Body)
			if diff := cmp.Diff(wantb, gotb); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func newBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func readBody(r io.ReadCloser) string {
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}
