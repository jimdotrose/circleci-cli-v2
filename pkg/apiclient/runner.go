package apiclient

import "net/url"

const runnerBase = "/runner"

// ListRunnerResourceClasses returns resource classes for a namespace.
func (c *Client) ListRunnerResourceClasses(namespace string) ([]RunnerResourceClass, error) {
	q := url.Values{}
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	var result struct {
		Items []RunnerResourceClass `json:"items"`
	}
	if err := c.get(runnerBase+"/resource", q, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// CreateRunnerResourceClass creates a new resource class.
func (c *Client) CreateRunnerResourceClass(resourceClass, description string) (*RunnerResourceClass, error) {
	body := map[string]string{
		"resource_class": resourceClass,
		"description":    description,
	}
	var rc RunnerResourceClass
	if err := c.post(runnerBase+"/resource", body, &rc); err != nil {
		return nil, err
	}
	return &rc, nil
}

// DeleteRunnerResourceClass deletes a resource class.
func (c *Client) DeleteRunnerResourceClass(resourceClass string) error {
	return c.delete(runnerBase + "/resource/" + resourceClass)
}

// ListRunnerTokens returns tokens for a resource class.
func (c *Client) ListRunnerTokens(resourceClass string) ([]RunnerToken, error) {
	q := url.Values{"resource-class": []string{resourceClass}}
	var result struct {
		Items []RunnerToken `json:"items"`
	}
	if err := c.get(runnerBase+"/token", q, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// CreateRunnerToken creates a new runner authentication token.
func (c *Client) CreateRunnerToken(resourceClass, nickname string) (*RunnerToken, error) {
	body := map[string]string{
		"resource_class": resourceClass,
		"nickname":       nickname,
	}
	var tok RunnerToken
	if err := c.post(runnerBase+"/token", body, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

// DeleteRunnerToken deletes a runner token by ID.
func (c *Client) DeleteRunnerToken(id string) error {
	return c.delete(runnerBase + "/token/" + id)
}

// ListRunnerInstances returns runner instances for a resource class.
func (c *Client) ListRunnerInstances(resourceClass string) ([]RunnerInstance, error) {
	q := url.Values{"resource-class": []string{resourceClass}}
	var result struct {
		Items []RunnerInstance `json:"items"`
	}
	if err := c.get(runnerBase+"/instances", q, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}
