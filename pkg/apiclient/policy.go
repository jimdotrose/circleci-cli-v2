package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const policyBase = "/orgs"

// PolicyBundlePush uploads a policy bundle directory to the CircleCI policy service.
func (c *Client) PolicyBundlePush(ownerID, bundleDir string, dryRun bool) (map[string]interface{}, error) {
	// Create a multipart form with all .rego files in the directory.
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".rego" {
			continue
		}
		fw, err := w.CreateFormFile("bundle", e.Name())
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(filepath.Join(bundleDir, e.Name()))
		if err != nil {
			return nil, err
		}
		fw.Write(data)
	}
	w.Close()

	path := fmt.Sprintf("%s/%s/policy-bundles", policyBase, ownerID)
	if dryRun {
		path += "?dry=true"
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1"+path, &body)
	if err != nil {
		return nil, apiError(err, "building request")
	}
	req.Header.Set("Circle-Token", c.token)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apiError(err, "executing request")
	}
	defer resp.Body.Close()
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// GetPolicyBundle fetches the current policy bundle.
func (c *Client) GetPolicyBundle(ownerID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.get(fmt.Sprintf("%s/%s/policy-bundles", policyBase, ownerID), nil, &result)
	return result, err
}

// GetPolicyDocument fetches a single policy document by name.
func (c *Client) GetPolicyDocument(ownerID, policyName string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.get(fmt.Sprintf("%s/%s/policy-bundles/%s", policyBase, ownerID, policyName), nil, &result)
	return result, err
}

// ListPolicyLogs returns decision logs for an owner.
func (c *Client) ListPolicyLogs(ownerID, after, before string, limit int, pageToken string) ([]PolicyLog, string, error) {
	q := url.Values{}
	if after != "" {
		q.Set("after", after)
	}
	if before != "" {
		q.Set("before", before)
	}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if pageToken != "" {
		q.Set("page-token", pageToken)
	}

	var page Page[PolicyLog]
	err := c.get(fmt.Sprintf("%s/%s/policy-decisions", policyBase, ownerID), q, &page)
	return page.Items, page.NextPageToken, err
}

// PolicyDecide evaluates a config against the policy bundle.
func (c *Client) PolicyDecide(ownerID, configPath, pipelineParamsJSON string, metaProjectID string) (*PolicyDecision, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"input": string(data),
	}
	if pipelineParamsJSON != "" {
		var params map[string]interface{}
		json.Unmarshal([]byte(pipelineParamsJSON), &params)
		body["metadata"] = map[string]interface{}{"pipeline_parameters": params}
	}
	if metaProjectID != "" {
		if m, ok := body["metadata"].(map[string]interface{}); ok {
			m["project_id"] = metaProjectID
		} else {
			body["metadata"] = map[string]interface{}{"project_id": metaProjectID}
		}
	}

	var dec PolicyDecision
	err = c.post(fmt.Sprintf("%s/%s/policy-decisions", policyBase, ownerID), body, &dec)
	return &dec, err
}

// GetPolicySettings returns policy settings for an owner.
func (c *Client) GetPolicySettings(ownerID string) (*PolicySettings, error) {
	var s PolicySettings
	err := c.get(fmt.Sprintf("%s/%s/policy-decisions/settings", policyBase, ownerID), nil, &s)
	return &s, err
}

// SetPolicySettings updates policy settings for an owner.
func (c *Client) SetPolicySettings(ownerID string, enabled bool) (*PolicySettings, error) {
	body := map[string]bool{"enabled": enabled}
	var s PolicySettings
	err := c.doWithBody("PATCH", fmt.Sprintf("/orgs/%s/policy-decisions/settings", ownerID), body, &s)
	return &s, err
}

// PolicyEval evaluates a local OPA bundle against config without uploading.
func (c *Client) PolicyEval(ownerID, bundlePath, configPath string) (*PolicyDecision, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	bundleData, err := os.ReadFile(bundlePath)
	if err != nil {
		return nil, err
	}
	body := map[string]string{
		"input":  string(data),
		"bundle": string(bundleData),
	}
	var dec PolicyDecision
	err = c.post(fmt.Sprintf("%s/%s/policy-decisions/eval", policyBase, ownerID), body, &dec)
	return &dec, err
}

// PolicyTest runs OPA tests against a policy bundle.
func (c *Client) PolicyTest(ownerID, bundleDir string) (map[string]interface{}, error) {
	body := map[string]string{"bundle_dir": bundleDir}
	var result map[string]interface{}
	err := c.post(fmt.Sprintf("%s/%s/policy-bundles/test", policyBase, ownerID), body, &result)
	return result, err
}

// PolicyDiff shows what would change by pushing a new bundle (dry run).
func (c *Client) PolicyDiff(ownerID, bundleDir string) (map[string]interface{}, error) {
	return c.PolicyBundlePush(ownerID, bundleDir, true)
}

// Ensure io is used via io.Discard reference.
var _ = io.Discard
