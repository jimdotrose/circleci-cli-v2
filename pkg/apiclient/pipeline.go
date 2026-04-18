package apiclient

import (
	"encoding/json"
	"net/url"
)

// ListPipelines returns pipelines for a project slug.
func (c *Client) ListPipelines(projectSlug, branch, pageToken string) ([]Pipeline, string, error) {
	path := "/project/" + projectSlug + "/pipeline"
	q := url.Values{}
	if branch != "" {
		q.Set("branch", branch)
	}
	if pageToken != "" {
		q.Set("page-token", pageToken)
	}

	var page Page[Pipeline]
	if err := c.get(path, q, &page); err != nil {
		return nil, "", err
	}
	return page.Items, page.NextPageToken, nil
}

// GetPipeline returns a pipeline by ID.
func (c *Client) GetPipeline(id string) (*Pipeline, error) {
	var p Pipeline
	if err := c.get("/pipeline/"+id, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// TriggerPipeline triggers a new pipeline for the project.
func (c *Client) TriggerPipeline(projectSlug, branch, tag string, parameters map[string]interface{}) (*TriggerPipelineResponse, error) {
	body := map[string]interface{}{}
	if branch != "" {
		body["branch"] = branch
	}
	if tag != "" {
		body["tag"] = tag
	}
	if len(parameters) > 0 {
		body["parameters"] = parameters
	}

	var resp TriggerPipelineResponse
	if err := c.post("/project/"+projectSlug+"/pipeline", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseParameters parses a JSON string into a map of pipeline parameters.
func ParseParameters(raw string) (map[string]interface{}, error) {
	if raw == "" {
		return nil, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	return m, nil
}
