package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// FeatureFlag represents a PostHog feature flag
type FeatureFlag struct {
	ID               int                    `json:"id"`
	Key              string                 `json:"key"`
	Name             string                 `json:"name"`
	Active           bool                   `json:"active"`
	Filters          map[string]interface{} `json:"filters"`
	CreatedAt        string                 `json:"created_at"`
	CreatedBy        interface{}            `json:"created_by"`
	Deleted          bool                   `json:"deleted"`
	EnsureExperience bool                   `json:"ensure_experience_continuity"`
}

// FlagsResponse represents the API response for feature flags list
type FlagsResponse struct {
	Next     *string       `json:"next"`
	Previous *string       `json:"previous"`
	Results  []FeatureFlag `json:"results"`
}

// ListFlags fetches all feature flags
func (c *Client) ListFlags(ctx context.Context) ([]FeatureFlag, error) {
	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	path := fmt.Sprintf("%s/feature_flags/", c.getProjectPath())

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var flagsResp FlagsResponse
	if err := json.Unmarshal(body, &flagsResp); err != nil {
		return nil, fmt.Errorf("failed to parse flags response: %w", err)
	}

	return flagsResp.Results, nil
}

// ToggleFlag updates a feature flag's active status
func (c *Client) ToggleFlag(ctx context.Context, flagID int, active bool) error {
	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	path := fmt.Sprintf("%s/feature_flags/%d/", c.getProjectPath(), flagID)

	data := map[string]interface{}{
		"active": active,
	}

	resp, err := c.patch(ctx, path, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetFlag fetches a single feature flag by ID
func (c *Client) GetFlag(ctx context.Context, flagID int) (*FeatureFlag, error) {
	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	path := fmt.Sprintf("%s/feature_flags/%d/", c.getProjectPath(), flagID)

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var flag FeatureFlag
	if err := json.Unmarshal(body, &flag); err != nil {
		return nil, fmt.Errorf("failed to parse flag response: %w", err)
	}

	return &flag, nil
}
