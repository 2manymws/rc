package testutil

// Copy from github.com/2manymws/rcutil

import (
	"bufio"
	"bytes"
	"net/http"
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
func decodeReqRes(c *cachedReqRes) (*http.Request, *http.Response, error) {
	reqb := bytes.NewReader(c.req)
	req, err := http.ReadRequest(bufio.NewReader(reqb))
	if err != nil {
		return nil, nil, err
	}
	resb := bytes.NewReader(c.res)
	res, err := http.ReadResponse(bufio.NewReader(resb), req)
	if err != nil {
		return nil, nil, err
	}
	return req, res, nil
}
