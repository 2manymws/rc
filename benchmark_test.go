package rc_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

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

	// Make cache
	wg := &sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/sleep/%d", proxy, i), nil)
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
			res.Body.Close()
		}()
	}

	rand.Seed(time.Now().UnixNano())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := rand.Intn(10000)
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/sleep/%d", proxy, i), nil)
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
			res.Body.Close()
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

	// Make cache
	wg := &sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/sleep/%d", proxy.URL, i), nil)
			if err != nil {
				b.Error(err)
				return
			}
			req.Host = hostname
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Error(err)
			}
			res.Body.Close()
		}()
	}

	rand.Seed(time.Now().UnixNano())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := rand.Intn(10000)
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/sleep/%d", proxy.URL, i), nil)
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
			res.Body.Close()
			got := res.StatusCode
			want := http.StatusOK
			if res.StatusCode != http.StatusOK {
				b.Errorf("got %v want %v", got, want)
			}
		}
	})
}

func TestContainer(t *testing.T) {
	hostname := "a.example.com"
	urlstr := testutil.NewUpstreamEchoNGINXServer(t, hostname)
	var upstreams = map[string]string{
		hostname: urlstr,
	}
	c := testuc.NewAllCache(t)
	m := rc.New(c)
	rl := testutil.NewRelayer(upstreams)
	r := rp.NewRouter(rl)
	proxy := httptest.NewServer(m(r))
	t.Cleanup(func() {
		proxy.Close()
	})

	{
		now := time.Now()
		req, err := http.NewRequest("GET", proxy.URL+"/sleep/hello", nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Host = hostname
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer res.Body.Close()
		after := time.Now()
		if after.Sub(now) < 1*time.Second {
			t.Fatal("sleep.js is not working")
		}
	}
	{
		req, err := http.NewRequest("GET", proxy.URL+"/sleep/hello", nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Host = hostname
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer res.Body.Close()
		if res.Header.Get("X-Cache") != "HIT" {
			t.Fatal("rp cache is not working")
		}
	}
}
