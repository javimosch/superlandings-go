package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RemoteClient handles HTTP requests to remote sl-cli daemon
type RemoteClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRemoteClient creates a new remote client
func NewRemoteClient(host string, port int) *RemoteClient {
	return &RemoteClient{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetStatus checks if remote daemon is running
func (c *RemoteClient) GetStatus() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

// ListSites lists all sites on remote daemon
func (c *RemoteClient) ListSites() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/sites")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

// GetSite gets site details from remote daemon
func (c *RemoteClient) GetSite(slug string) (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/sites/" + slug)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

// ListVersions lists versions for a site
func (c *RemoteClient) ListVersions(slug string) (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/sites/" + slug + "/versions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

// SyncSite triggers sync operation on remote daemon
func (c *RemoteClient) SyncSite(slug string, payload map[string]interface{}) (map[string]interface{}, error) {
	return c.postJSON("/api/sites/"+slug+"/sync", payload)
}

func (c *RemoteClient) parseResponse(resp *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// Try to parse as array first
	var arr []interface{}
	if err := json.Unmarshal(body, &arr); err == nil {
		// It's an array, wrap it in a map
		return map[string]interface{}{"sites": arr}, nil
	}
	
	// Otherwise parse as object
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return result, nil
}

func (c *RemoteClient) postJSON(path string, payload map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}