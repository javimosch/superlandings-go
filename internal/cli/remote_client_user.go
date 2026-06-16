package cli

// User remote methods
func (c *RemoteClient) ListUsers() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/users", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

func (c *RemoteClient) CreateUser(email, password, role string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"email":    email,
		"password": password,
		"role":     role,
	}
	return c.postJSON("/api/users", payload)
}

func (c *RemoteClient) SetUserPassword(email, password string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"password": password,
	}
	return c.postJSON("/api/users/"+email+"/password", payload)
}

func (c *RemoteClient) GrantSiteAccess(siteSlug, email, role string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"site_slug": siteSlug,
		"email":     email,
		"role":      role,
	}
	return c.postJSON("/api/users/grant", payload)
}

// Site admin remote methods
func (c *RemoteClient) CreateSiteAdminToken(siteSlug string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("POST", "/api/sites/"+siteSlug+"/admin", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

func (c *RemoteClient) GetSiteAdminToken(siteSlug string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/sites/"+siteSlug+"/admin", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

func (c *RemoteClient) RotateSiteAdminToken(siteSlug string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("PUT", "/api/sites/"+siteSlug+"/admin", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

func (c *RemoteClient) RevokeSiteAdminToken(siteSlug string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("DELETE", "/api/sites/"+siteSlug+"/admin", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}