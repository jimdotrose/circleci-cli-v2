package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
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

// GetJobLogs streams step output for a job to w.
// It fetches the list of steps, then retrieves and writes each action's output.
func (c *Client) GetJobLogs(projectSlug string, jobNumber int, w io.Writer) error {
	var result struct {
		Items []JobStep `json:"items"`
	}
	if err := c.get("/project/"+projectSlug+"/job/"+itoa(jobNumber)+"/steps", nil, &result); err != nil {
		return err
	}
	for _, step := range result.Items {
		for _, action := range step.Actions {
			if action.OutputURL == "" {
				continue
			}
			if err := c.writeStepOutput(action.OutputURL, w); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeStepOutput fetches a pre-signed step output URL and writes each log
// message to w. Step output URLs are self-authenticating and require no token.
func (c *Client) writeStepOutput(outputURL string, w io.Writer) error {
	resp, err := c.httpClient.Get(outputURL)
	if err != nil {
		return apiError(err, "fetching step output")
	}
	defer resp.Body.Close()
	if err := checkStatus(resp); err != nil {
		return err
	}
	var msgs []StepOutputMessage
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		return apiError(err, "decoding step output")
	}
	for _, m := range msgs {
		fmt.Fprint(w, m.Message)
	}
	return nil
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
