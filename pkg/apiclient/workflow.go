package apiclient

import "net/url"

// ListWorkflows returns workflows for a pipeline.
func (c *Client) ListWorkflows(pipelineID, pageToken string) ([]Workflow, string, error) {
	q := url.Values{}
	if pageToken != "" {
		q.Set("page-token", pageToken)
	}

	var page Page[Workflow]
	if err := c.get("/pipeline/"+pipelineID+"/workflow", q, &page); err != nil {
		return nil, "", err
	}
	return page.Items, page.NextPageToken, nil
}

// GetWorkflow returns a workflow by ID.
func (c *Client) GetWorkflow(id string) (*Workflow, error) {
	var w Workflow
	if err := c.get("/workflow/"+id, nil, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// CancelWorkflow cancels a running workflow.
func (c *Client) CancelWorkflow(id string) error {
	return c.post("/workflow/"+id+"/cancel", nil, nil)
}

// RerunWorkflow reruns a workflow, optionally from failed jobs.
func (c *Client) RerunWorkflow(id string, fromFailed bool) error {
	body := map[string]interface{}{
		"enable_ssh": false,
		"from_failed": fromFailed,
	}
	return c.post("/workflow/"+id+"/rerun", body, nil)
}
