package rc_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k1LoW/rc"
	testuc "github.com/k1LoW/rc/testutil"
	"github.com/k1LoW/rp"
	"github.com/k1LoW/rp/testutil"
)

func BenchmarkNGINXCache(b *testing.B) {
	hostname := "a.example.com"
	_ = testutil.NewUpstreamEchoNGINXServer(b, hostname)
	var upstreams = map[string]string{
		hostname: fmt.Sprintf("http://%s:80", hostname),
	}
	proxy := testutil.NewReverseProxyNGINXServer(b, "r.example.com", upstreams)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest("GET", proxy+"sleep", nil)
			if err != nil {
				b.Error(err)
				return
			}
			req.Host = hostname
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Error(err)
				return
			}
			got := res.StatusCode
			want := http.StatusOK
			if res.StatusCode != http.StatusOK {
				b.Errorf("got %v want %v", got, want)
			}
		}
	})
}

func BenchmarkRC(b *testing.B) {
	hostname := "a.example.com"
	urlstr := testutil.NewUpstreamEchoNGINXServer(b, hostname)
	var upstreams = map[string]string{
		hostname: urlstr,
	}
	c := testuc.NewAllCache(b)
	m := rc.New(c)
	rl := testutil.NewRelayer(upstreams)
	r := rp.NewRouter(rl)
	proxy := httptest.NewServer(m(r))
	b.Cleanup(func() {
		proxy.Close()
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest("GET", proxy.URL+"/sleep", nil)
			if err != nil {
				b.Error(err)
				return
			}
			req.Host = hostname
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Error(err)
				return
			}
			got := res.StatusCode
			want := http.StatusOK
			if res.StatusCode != http.StatusOK {
				b.Errorf("got %v want %v", got, want)
			}
		}
	})
}
