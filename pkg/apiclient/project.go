package apiclient

import "net/url"

// GetProject returns a project by slug.
func (c *Client) GetProject(slug string) (*Project, error) {
	var p Project
	if err := c.get("/project/"+slug, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// FollowProject follows a project, enabling builds.
func (c *Client) FollowProject(slug string) (*Project, error) {
	var p Project
	if err := c.post("/project/"+slug+"/follow", nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListProjects returns projects the authenticated user follows.
func (c *Client) ListProjects(pageToken string) ([]Project, string, error) {
	q := url.Values{}
	if pageToken != "" {
		q.Set("page-token", pageToken)
	}
	var page Page[Project]
	if err := c.get("/me/collaborations", q, &page); err != nil {
		// /me/collaborations is a different shape; fall back to array decode
		var arr []Project
		if err2 := c.get("/me/collaborations", q, &arr); err2 != nil {
			return nil, "", err
		}
		return arr, "", nil
	}
	return page.Items, page.NextPageToken, nil
}

// ListEnvVars returns environment variables for a project (values redacted).
func (c *Client) ListEnvVars(slug string) ([]EnvVar, error) {
	var page Page[EnvVar]
	if err := c.get("/project/"+slug+"/envvar", nil, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}

// GetEnvVar returns a single environment variable (value is always "xxxx").
func (c *Client) GetEnvVar(slug, name string) (*EnvVar, error) {
	var ev EnvVar
	if err := c.get("/project/"+slug+"/envvar/"+name, nil, &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

// SetEnvVar creates or updates a project environment variable.
func (c *Client) SetEnvVar(slug, name, value string) error {
	body := map[string]string{"name": name, "value": value}
	return c.post("/project/"+slug+"/envvar", body, nil)
}

// DeleteEnvVar deletes a project environment variable.
func (c *Client) DeleteEnvVar(slug, name string) error {
	return c.delete("/project/" + slug + "/envvar/" + name)
}
