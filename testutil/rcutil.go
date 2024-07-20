package testutil

// Copy from github.com/2manymws/rcutil

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"testing"
)

type cachedReqRes struct {
	req []byte
	res []byte
}

// encodeReqRes encodes http.Request and http.Response.
func encodeReqRes(req *http.Request, res *http.Response) (*cachedReqRes, error) {
	reqb := &bytes.Buffer{}
	if err := req.Write(reqb); err != nil {
		return nil, err
	}
	resb := &bytes.Buffer{}
	if err := res.Write(resb); err != nil {
		return nil, err
	}

	c := &cachedReqRes{
		req: reqb.Bytes(),
		res: resb.Bytes(),
	}

	return c, nil
}

// decodeReqRes decodes to http.Request and http.Response.
func decodeReqRes(t testing.TB, c *cachedReqRes) (*http.Request, *http.Response, error) {
	reqb := bytes.NewReader(c.req)
	req, err := http.ReadRequest(bufio.NewReader(reqb))
	if err != nil {
		return nil, nil, err
	}
	req.Body = NewCloseChecker(t, req.Body)
	resb := bytes.NewReader(c.res)
	res, err := http.ReadResponse(bufio.NewReader(resb), req)
	if err != nil {
		return nil, nil, err
	}
	res.Body = NewCloseChecker(t, res.Body)
	return req, res, nil
}

type CloseChecker struct {
	rc     io.ReadCloser
	closed bool
}

func NewCloseChecker(t testing.TB, rc io.ReadCloser) *CloseChecker {
	r := &CloseChecker{
		rc: rc,
	}
	t.Cleanup(func() {
		if !r.closed {
			t.Errorf("CloseChecker: not closed")
		}
	})
	return r
}

func (r *CloseChecker) Read(p []byte) (n int, err error) {
	return r.rc.Read(p)
}

func (r *CloseChecker) Close() error {
	r.closed = true
	return r.rc.Close()
}
