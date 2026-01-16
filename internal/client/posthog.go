package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	debug       bool
	projectID   int
}

// New creates a new PostHog API client
func New(cfg *config.Config) *Client {
	return &Client{
		apiKey:      cfg.ProjectAPIKey,
		instanceURL: cfg.InstanceURL,
		debug:       cfg.Debug,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.instanceURL, path)

	// Read body for debugging (if present)
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = io.ReadAll(body)
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// PostHog uses personal API key in header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	if c.debug {
		log.Printf("[DEBUG] Request: %s %s", method, url)
		maskedKey := "phx_****" + c.apiKey[len(c.apiKey)-4:]
		log.Printf("[DEBUG] Authorization: Bearer %s", maskedKey)
		if len(bodyBytes) > 0 {
			log.Printf("[DEBUG] Request Body: %s", string(bodyBytes))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if c.debug {
			log.Printf("[DEBUG] Request failed: %v", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if c.debug {
		log.Printf("[DEBUG] Response Status: %d %s", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if c.debug {
			log.Printf("[DEBUG] Response Body: %s", string(body))
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// For successful responses, log body in debug mode
	if c.debug {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			log.Printf("[DEBUG] Response Body: %s", string(bodyBytes))
			// Recreate the response body so it can be read again
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
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

// ProjectInfo represents basic project information
type ProjectInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// InitializeProject fetches and caches the current project ID
func (c *Client) InitializeProject(ctx context.Context) error {
	// Try to get user info which includes project information
	resp, err := c.get(ctx, "/api/users/@me/")
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse to get the current project ID
	var userInfo struct {
		Team struct {
			ID int `json:"id"`
		} `json:"team"`
	}
	if err := json.Unmarshal(bodyBytes, &userInfo); err != nil {
		return fmt.Errorf("failed to parse user info: %w", err)
	}

	if userInfo.Team.ID == 0 {
		return fmt.Errorf("no team/project found in user info")
	}

	c.projectID = userInfo.Team.ID
	if c.debug {
		log.Printf("[DEBUG] Project ID initialized: %d", c.projectID)
	}

	return nil
}

// TestConnection verifies the API credentials work
func (c *Client) TestConnection(ctx context.Context) error {
	// Initialize project ID
	if err := c.InitializeProject(ctx); err != nil {
		return err
	}

	// Test the project endpoint with the actual project ID
	path := fmt.Sprintf("/api/projects/%d/", c.projectID)
	resp, err := c.get(ctx, path)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetProjectID returns the cached project ID
func (c *Client) GetProjectID() int {
	return c.projectID
}

// getProjectPath returns the project API path prefix
func (c *Client) getProjectPath() string {
	if c.projectID > 0 {
		return fmt.Sprintf("/api/projects/%d", c.projectID)
	}
	return "/api/projects/@current"
}
