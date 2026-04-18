package apiclient

import (
	"fmt"
	"net/url"
)

// ListJobs returns jobs for a workflow.
func (c *Client) ListJobs(workflowID, pageToken string) ([]Job, string, error) {
	q := url.Values{}
	if pageToken != "" {
		q.Set("page-token", pageToken)
	}

	var page Page[Job]
	if err := c.get("/workflow/"+workflowID+"/job", q, &page); err != nil {
		return nil, "", err
	}
	return page.Items, page.NextPageToken, nil
}

// GetJob returns a job by project slug and job number.
func (c *Client) GetJob(projectSlug string, jobNumber int) (*Job, error) {
	var j Job
	if err := c.get("/project/"+projectSlug+"/job/"+itoa(jobNumber), nil, &j); err != nil {
		return nil, err
	}
	return &j, nil
}

// CancelJob cancels a running job.
func (c *Client) CancelJob(projectSlug string, jobNumber int) error {
	return c.post("/project/"+projectSlug+"/job/"+itoa(jobNumber)+"/cancel", nil, nil)
}

// ListArtifacts returns artifacts for a job.
func (c *Client) ListArtifacts(projectSlug string, jobNumber int, pageToken string) ([]Artifact, string, error) {
	q := url.Values{}
	if pageToken != "" {
		q.Set("page-token", pageToken)
	}

	var page Page[Artifact]
	if err := c.get("/project/"+projectSlug+"/"+itoa(jobNumber)+"/artifacts", q, &page); err != nil {
		return nil, "", err
	}
	return page.Items, page.NextPageToken, nil
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
