package apiclient

import "encoding/json"

func jsonUnmarshal(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// CompileConfig calls the CircleCI API to validate and process a config.
// orgID is optional; pass "" to skip.
func (c *Client) CompileConfig(configYAML, pipelineParamsJSON, orgID string) (*CompileResponse, error) {
	body := map[string]interface{}{
		"config": configYAML,
	}
	if orgID != "" {
		body["org_id"] = orgID
	}
	if pipelineParamsJSON != "" {
		var params map[string]interface{}
		if err := jsonUnmarshal(pipelineParamsJSON, &params); err == nil {
			body["pipeline_parameters"] = params
		}
	}

	var resp CompileResponse
	if err := c.post("/compile-config-with-defaults", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Me returns basic info about the authenticated user (used by diagnostic).
func (c *Client) Me() (map[string]interface{}, error) {
	var me map[string]interface{}
	if err := c.get("/me", nil, &me); err != nil {
		return nil, err
	}
	return me, nil
}
