package rc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type CacheResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func ResponseToBytes(res *http.Response) ([]byte, error) {
	c := &CacheResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	c.Body = b
	cb, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return cb, nil
}

func BytesToResponse(b []byte) (*http.Response, error) {
	c := &CacheResponse{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	res := &http.Response{
		StatusCode: c.StatusCode,
		Header:     c.Header,
		Body:       io.NopCloser(bytes.NewReader(c.Body)),
	}
	return res, nil
}
