package apiclient

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// RawRequest performs an authenticated HTTP request to the given API path and
// returns the raw response body. body may be nil for GET/DELETE requests.
// extraHeaders are merged with the default Circle-Token and Content-Type headers.
func (c *Client) RawRequest(method, path string, body map[string]interface{}, extraHeaders map[string]string) ([]byte, error) {
	u := c.baseURL + "/api/v2" + path

	var bodyReader io.Reader
	if len(body) > 0 {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, apiError(err, "encoding request body")
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, apiError(err, "building request")
	}
	req.Header.Set("Circle-Token", c.token)
	req.Header.Set("Accept", "application/json")
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apiError(err, "executing request")
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}
