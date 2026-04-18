package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

const defaultTimeout = 30 * time.Second

// Client is a thin CircleCI REST API v2 client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// New returns a Client configured to speak to baseURL with the given token.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Page holds a single page of results from a list endpoint.
type Page[T any] struct {
	Items         []T    `json:"items"`
	NextPageToken string `json:"next_page_token"`
}

// get performs a GET request and decodes JSON into out.
func (c *Client) get(path string, query url.Values, out interface{}) error {
	u := c.baseURL + "/api/v2" + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return apiError(err, "building request")
	}
	req.Header.Set("Circle-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apiError(err, "executing request")
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// post performs a POST request with a JSON body and decodes the response into out.
// out may be nil if no response body is expected.
func (c *Client) post(path string, body interface{}, out interface{}) error {
	return c.doWithBody(http.MethodPost, path, body, out)
}

// delete performs a DELETE request.
func (c *Client) delete(path string) error {
	u := c.baseURL + "/api/v2" + path
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return apiError(err, "building request")
	}
	req.Header.Set("Circle-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apiError(err, "executing request")
	}
	defer resp.Body.Close()
	return checkStatus(resp)
}

func (c *Client) doWithBody(method, path string, body interface{}, out interface{}) error {
	pr, pw := io.Pipe()
	go func() {
		enc := json.NewEncoder(pw)
		pw.CloseWithError(enc.Encode(body))
	}()

	req, err := http.NewRequest(method, c.baseURL+"/api/v2"+path, pr)
	if err != nil {
		return apiError(err, "building request")
	}
	req.Header.Set("Circle-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apiError(err, "executing request")
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return err
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// checkStatus maps non-2xx responses to CLIErrors with appropriate exit codes.
func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	msg := string(body)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return cierrors.New(
			"AUTH_REQUIRED",
			"Authentication failed",
			"Your API token is missing or invalid.",
			cierrors.ExitAuthError,
		).WithSuggestions(
			"Run: circleci auth login",
			"Or set the CIRCLECI_TOKEN environment variable",
		).WithRef("https://circleci.com/docs/local-cli/")

	case http.StatusForbidden:
		return cierrors.New(
			"FORBIDDEN",
			"Access denied",
			fmt.Sprintf("You do not have permission to perform this action. %s", msg),
			cierrors.ExitAuthError,
		)

	case http.StatusNotFound:
		return cierrors.New(
			"NOT_FOUND",
			"Resource not found",
			fmt.Sprintf("The requested resource does not exist. %s", msg),
			cierrors.ExitNotFound,
		).WithSuggestions(
			"Use a list command to see available resources (e.g. 'circleci context list')",
			"Check that the ID or slug is correct",
		)

	default:
		return cierrors.New(
			"API_ERROR",
			fmt.Sprintf("API error %d", resp.StatusCode),
			fmt.Sprintf("The CircleCI API returned an error: %s", msg),
			cierrors.ExitAPIError,
		)
	}
}

func apiError(err error, context string) error {
	return cierrors.New(
		"API_ERROR",
		"Request failed",
		fmt.Sprintf("Error %s: %v", context, err),
		cierrors.ExitAPIError,
	)
}
