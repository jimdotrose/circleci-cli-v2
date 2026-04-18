package apiclient

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

// debugTransport logs every request and response when debug is enabled.
type debugTransport struct {
	base   http.RoundTripper
	output io.Writer
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Fprintf(t.output, "→ %s\n", dump)

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	dump, _ = httputil.DumpResponse(resp, true)
	fmt.Fprintf(t.output, "← %s\n", dump)
	return resp, nil
}

// NewWithDebug returns a Client that logs HTTP traffic to w when debug is true.
func NewWithDebug(baseURL, token string, debug bool, w io.Writer) *Client {
	c := New(baseURL, token)
	if debug {
		c.httpClient.Transport = &debugTransport{
			base:   http.DefaultTransport,
			output: w,
		}
	}
	return c
}
