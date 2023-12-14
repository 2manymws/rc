package testutil

// Copy from github.com/k1LoW/rcutil

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

type cachedReqRes struct {
	Method    string      `json:"method"`
	URL       string      `json:"url"`
	ReqHeader http.Header `json:"req_header"`
	ReqBody   []byte      `json:"req_body"`

	StatusCode int         `json:"status_code"`
	ResHeader  http.Header `json:"res_header"`
	ResBody    []byte      `json:"res_body"`
}

// encodeReqRes encodes http.Request and http.Response.
func encodeReqRes(req *http.Request, res *http.Response) (*cachedReqRes, error) {
	c := &cachedReqRes{
		Method:    req.Method,
		URL:       req.URL.String(),
		ReqHeader: req.Header,

		StatusCode: res.StatusCode,
		ResHeader:  res.Header,
	}
	{
		// FIXME: Use stream
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		defer req.Body.Close()
		c.ReqBody = b
	}

	{
		// FIXME: Use stream
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		c.ResBody = b
	}

	return c, nil
}

// decodeReqRes decodes to http.Request and http.Response.
func decodeReqRes(c *cachedReqRes) (*http.Request, *http.Response, error) {
	u, err := url.Parse(c.URL)
	if err != nil {
		return nil, nil, err
	}
	req := &http.Request{
		Method: c.Method,
		URL:    u,
		Header: c.ReqHeader,
		Body:   io.NopCloser(bytes.NewReader(c.ReqBody)),
	}
	res := &http.Response{
		Status:     http.StatusText(c.StatusCode),
		StatusCode: c.StatusCode,
		Header:     c.ResHeader,
		Body:       io.NopCloser(bytes.NewReader(c.ResBody)),
	}
	return req, res, nil
}
