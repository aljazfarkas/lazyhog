package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/config"
)

// Client is a PostHog API client wrapper
type Client struct {
	apiKey      string
	instanceURL string
	httpClient  *http.Client
}

// New creates a new PostHog API client
func New(cfg *config.Config) *Client {
	return &Client{
		apiKey:      cfg.ProjectAPIKey,
		instanceURL: cfg.InstanceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.instanceURL, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// PostHog uses personal API key in header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// get performs a GET request
func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	return c.doRequest(ctx, "GET", path, nil)
}

// post performs a POST request
func (c *Client) post(ctx context.Context, path string, data interface{}) (*http.Response, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}
	return c.doRequest(ctx, "POST", path, body)
}

// patch performs a PATCH request
func (c *Client) patch(ctx context.Context, path string, data interface{}) (*http.Response, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}
	return c.doRequest(ctx, "PATCH", path, body)
}

// TestConnection verifies the API credentials work
func (c *Client) TestConnection(ctx context.Context) error {
	resp, err := c.get(ctx, "/api/projects/@current/")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
