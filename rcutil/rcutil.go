package rcutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type cacheResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// ResponseToBytes converts http.Response to []byte.
func ResponseToBytes(res *http.Response) ([]byte, error) {
	c := &cacheResponse{
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

// BytesToResponse converts []byte to http.Response.
func BytesToResponse(b []byte) (*http.Response, error) {
	c := &cacheResponse{}
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
