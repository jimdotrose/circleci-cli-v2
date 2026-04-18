package apiclient

import "net/url"

// ListContexts returns all contexts for the given owner (org-slug or org-id).
func (c *Client) ListContexts(ownerSlug, ownerType, pageToken string) ([]Context, string, error) {
	q := url.Values{}
	if ownerSlug != "" {
		q.Set("owner-slug", ownerSlug)
	}
	if ownerType != "" {
		q.Set("owner-type", ownerType)
	}
	if pageToken != "" {
		q.Set("page-token", pageToken)
	}

	var page Page[Context]
	if err := c.get("/context", q, &page); err != nil {
		return nil, "", err
	}
	return page.Items, page.NextPageToken, nil
}

// GetContext returns a context by ID.
func (c *Client) GetContext(id string) (*Context, error) {
	var ctx Context
	if err := c.get("/context/"+id, nil, &ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}

// CreateContext creates a new context for the given owner.
func (c *Client) CreateContext(name, ownerID, ownerType string) (*Context, error) {
	body := map[string]interface{}{
		"name": name,
		"owner": map[string]string{
			"id":   ownerID,
			"type": ownerType,
		},
	}
	var ctx Context
	if err := c.post("/context", body, &ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}

// DeleteContext deletes a context by ID.
func (c *Client) DeleteContext(id string) error {
	return c.delete("/context/" + id)
}

// ListContextVariables returns the variables for a context.
func (c *Client) ListContextVariables(contextID, pageToken string) ([]ContextVariable, string, error) {
	q := url.Values{}
	if pageToken != "" {
		q.Set("page-token", pageToken)
	}

	var page Page[ContextVariable]
	if err := c.get("/context/"+contextID+"/environment-variable", q, &page); err != nil {
		return nil, "", err
	}
	return page.Items, page.NextPageToken, nil
}

// SetContextVariable creates or updates a context variable.
func (c *Client) SetContextVariable(contextID, name, value string) error {
	body := map[string]string{"value": value}
	return c.doWithBody("PUT", "/context/"+contextID+"/environment-variable/"+name, body, nil)
}

// RemoveContextVariable deletes a context variable.
func (c *Client) RemoveContextVariable(contextID, name string) error {
	return c.delete("/context/" + contextID + "/environment-variable/" + name)
}
