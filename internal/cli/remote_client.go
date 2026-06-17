package cli

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/javimosch/superlandings-go/internal/config"
)

// RemoteClient handles HTTP requests to remote sl-cli daemon
type RemoteClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// NewRemoteClient creates a new remote client from host:port
func NewRemoteClient(host string, port int) *RemoteClient {
	return &RemoteClient{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewRemoteClientFromTarget creates a new remote client from a target name
func NewRemoteClientFromTarget(targetName string) (*RemoteClient, error) {
	// Check if target looks like host:port
	if strings.Contains(targetName, ":") {
		parts := strings.Split(targetName, ":")
		host := parts[0]
		port := 3100
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &port)
		}
		return NewRemoteClient(host, port), nil
	}
	
	// Load target from config
	target, err := config.GetTarget(targetName)
	if err != nil {
		return nil, fmt.Errorf("target '%s' not found: %w", targetName, err)
	}
	
	client := NewRemoteClient(target.Host, target.Port)
	client.authToken = target.AuthToken
	return client, nil
}

func (c *RemoteClient) GetStatus() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

func (c *RemoteClient) ListSites() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/sites", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

func (c *RemoteClient) GetSite(slug string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/sites/"+slug, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

func (c *RemoteClient) ListVersions(slug string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/sites/"+slug+"/versions", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

func (c *RemoteClient) CreateVersion(siteSlug string, version, comment, author string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"version": version,
		"comment": comment,
		"author":  author,
	}
	return c.postJSON("/api/sites/"+siteSlug+"/versions", payload)
}

func (c *RemoteClient) SwitchVersion(siteSlug, version string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"version": version,
	}
	return c.postJSON("/api/sites/"+siteSlug+"/versions/switch", payload)
}

func (c *RemoteClient) WriteFile(siteSlug, version, file, content string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"version": version,
		"file":     file,
		"content":  content,
	}
	return c.postJSON("/api/sites/"+siteSlug+"/write", payload)
}

// WriteBatch writes multiple files to a site in a single request
func (c *RemoteClient) WriteBatch(siteSlug, version string, files []map[string]string) (map[string]interface{}, error) {
	var fileList []map[string]string
	for _, f := range files {
		fileList = append(fileList, map[string]string{
			"file":    f["file"],
			"content": f["content"],
		})
	}
	payload := map[string]interface{}{
		"version": version,
		"files":   fileList,
	}
	return c.postJSON("/api/sites/"+siteSlug+"/write-batch", payload)
}

func (c *RemoteClient) UploadAsset(siteSlug, assetPath string, data []byte) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"path": assetPath,
		"data": base64.StdEncoding.EncodeToString(data),
	}
	return c.postJSON("/api/sites/"+siteSlug+"/upload", payload)
}

func (c *RemoteClient) ListAssets(siteSlug string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/sites/"+siteSlug+"/assets", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return c.parseResponse(resp)
}

func (c *RemoteClient) RemoveAsset(siteSlug, assetPath string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("DELETE", "/api/sites/"+siteSlug+"/assets/"+assetPath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return c.parseResponse(resp)
}

func (c *RemoteClient) SyncSite(slug string, payload map[string]interface{}) (map[string]interface{}, error) {
	return c.postJSON("/api/sites/"+slug+"/sync", payload)
}

func (c *RemoteClient) ListDNS(siteSlug string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/sites/"+siteSlug+"/dns", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}

func (c *RemoteClient) SetupDNS(siteSlug string, domain, ip string, traefik bool) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"domain":  domain,
		"ip":      ip,
		"traefik": traefik,
	}
	return c.postJSON("/api/sites/"+siteSlug+"/dns/setup", payload)
}

func (c *RemoteClient) RemoveDNS(siteSlug string) (map[string]interface{}, error) {
	return c.postJSON("/api/sites/"+siteSlug+"/dns/remove", map[string]interface{}{})
}

func (c *RemoteClient) makeRequest(method, path string, body []byte) (*http.Response, error) {
	url := c.baseURL + path
	
	var req *http.Request
	var err error
	
	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Add auth token if configured
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	
	return c.httpClient.Do(req)
}

func (c *RemoteClient) parseResponse(resp *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Check for API-level error
	if resp.StatusCode >= 400 {
		errMsg := "remote API error"
		if e, ok := result["error"].(string); ok && e != "" {
			errMsg = e
		}
		return result, fmt.Errorf("%s (HTTP %d)", errMsg, resp.StatusCode)
	}

	return result, nil
}

func (c *RemoteClient) postJSON(path string, payload map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.makeRequest("POST", path, jsonData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return c.parseResponse(resp)
}