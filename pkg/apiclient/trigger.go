package apiclient

import "net/url"

// ListScheduledTriggers returns scheduled pipeline triggers for a project.
func (c *Client) ListScheduledTriggers(projectSlug string) ([]ScheduledTrigger, error) {
	var page Page[ScheduledTrigger]
	if err := c.get("/project/"+projectSlug+"/schedule", nil, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}

// CreateScheduledTrigger creates a new scheduled pipeline trigger.
func (c *Client) CreateScheduledTrigger(projectSlug, name, description, actorID string, timetable map[string]interface{}, parameters map[string]interface{}) (*ScheduledTrigger, error) {
	body := map[string]interface{}{
		"name":        name,
		"description": description,
		"timetable":   timetable,
		"actor":       map[string]string{"id": actorID},
	}
	if len(parameters) > 0 {
		body["parameters"] = parameters
	}
	var st ScheduledTrigger
	if err := c.post("/project/"+projectSlug+"/schedule", body, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// CreateNamespace creates a new orb namespace.
// ownerID is the organization ID; ownerType is "organization" or "account".
func (c *Client) CreateNamespace(name, ownerID, ownerType string) (map[string]interface{}, error) {
	q := url.Values{}
	body := map[string]interface{}{
		"name": name,
		"owner": map[string]string{
			"id":   ownerID,
			"type": ownerType,
		},
	}
	_ = q
	var result map[string]interface{}
	if err := c.post("/orb/ns", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
