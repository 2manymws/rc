package rc_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/k1LoW/rc"
	"github.com/k1LoW/rc/testutil"
)

func TestResponseToBytesToResponse(t *testing.T) {
	tests := []struct {
		res  *http.Response
		want *http.Response
	}{
		{
			res:  &http.Response{Body: testutil.NewBody("")},
			want: &http.Response{Body: testutil.NewBody("")},
		},
		{
			res:  &http.Response{Status: http.StatusText(http.StatusOK), StatusCode: http.StatusOK, Body: testutil.NewBody("")},
			want: &http.Response{StatusCode: http.StatusOK, Body: testutil.NewBody("")},
		},
		{
			res:  &http.Response{Header: http.Header{"X-Cache": []string{"HIT"}, "X-Hello": []string{"World"}}, Body: testutil.NewBody("")},
			want: &http.Response{Header: http.Header{"X-Cache": []string{"HIT"}, "X-Hello": []string{"World"}}, Body: testutil.NewBody("")},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b, err := rc.ResponseToBytes(tt.res)
			if err != nil {
				t.Fatal(err)
			}
			got, err := rc.BytesToResponse(b)
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
			gotb := testutil.ReadBody(got.Body)
			wantb := testutil.ReadBody(tt.want.Body)
			if diff := cmp.Diff(wantb, gotb); diff != "" {
				t.Error(diff)
			}
		})
	}
}
