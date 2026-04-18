package apiclient

import "net/url"

// ListOrbs returns orbs from the registry, optionally filtered by namespace.
// When private is true the request includes privately accessible orbs.
// sort may be "popularity", "latest", or "alphabetical".
func (c *Client) ListOrbs(namespace string, private bool, sort string) ([]Orb, error) {
	q := url.Values{}
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	if private {
		q.Set("mine", "true")
	}
	if sort != "" {
		q.Set("sort", sort)
	}

	var page Page[Orb]
	if err := c.get("/orb", q, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}

// GetOrb returns details for the orb identified by "namespace/name".
func (c *Client) GetOrb(name string) (*Orb, error) {
	var orb Orb
	if err := c.get("/orb/"+name, nil, &orb); err != nil {
		return nil, err
	}
	return &orb, nil
}

// ValidateOrb validates an orb YAML string and returns the result.
func (c *Client) ValidateOrb(config string) (*OrbValidateResponse, error) {
	body := map[string]string{"config": config}
	var resp OrbValidateResponse
	if err := c.post("/orb/validate", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PublishOrb publishes a new orb version from YAML source.
// orbName is "namespace/name"; version is a semver or dev label.
func (c *Client) PublishOrb(orbName, version, config string) (*OrbVersion, error) {
	body := map[string]string{"config": config}
	var resp struct {
		Orb OrbVersion `json:"orb"`
	}
	if err := c.post("/orb/"+orbName+"/"+version, body, &resp); err != nil {
		return nil, err
	}
	return &resp.Orb, nil
}

// PromoteOrb promotes a dev orb version to a semantic release.
// segment must be "major", "minor", or "patch".
func (c *Client) PromoteOrb(orbName, devVersion, segment string) (*OrbVersion, error) {
	body := map[string]string{
		"dev_version":  devVersion,
		"release_type": segment,
	}
	var resp struct {
		Orb OrbVersion `json:"orb"`
	}
	if err := c.post("/orb/"+orbName+"/promote", body, &resp); err != nil {
		return nil, err
	}
	return &resp.Orb, nil
}

// SearchOrbs searches the orb registry for orbs matching query.
func (c *Client) SearchOrbs(query string) ([]Orb, error) {
	q := url.Values{}
	q.Set("query", query)
	var page Page[Orb]
	if err := c.get("/orb", q, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}
