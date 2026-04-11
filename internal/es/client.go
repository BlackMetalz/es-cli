package es

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

func NewClient(baseURL, username, password string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

func (c *Client) do(method, path string, body string) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ES error (HTTP %d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (c *Client) Get(path string) ([]byte, error) {
	return c.do(http.MethodGet, path, "")
}

func (c *Client) Put(path string, body string) ([]byte, error) {
	return c.do(http.MethodPut, path, body)
}

func (c *Client) Post(path string, body string) ([]byte, error) {
	return c.do(http.MethodPost, path, body)
}

func (c *Client) Delete(path string) ([]byte, error) {
	return c.do(http.MethodDelete, path, "")
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) Username() string {
	return c.username
}

type ClusterInfo struct {
	Name    string
	Version string
	Health  string
}

func (c *Client) GetClusterInfo() (*ClusterInfo, error) {
	info := &ClusterInfo{}

	data, err := c.Get("/")
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if name, ok := raw["cluster_name"].(string); ok {
		info.Name = name
	}
	if ver, ok := raw["version"].(map[string]interface{}); ok {
		if num, ok := ver["number"].(string); ok {
			info.Version = num
		}
	}

	healthData, err := c.Get("/_cluster/health")
	if err == nil {
		var healthRaw map[string]interface{}
		if json.Unmarshal(healthData, &healthRaw) == nil {
			if status, ok := healthRaw["status"].(string); ok {
				info.Health = status
			}
		}
	}

	return info, nil
}
