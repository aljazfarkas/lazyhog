package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Event represents a PostHog event
type Event struct {
	ID         string                 `json:"id"`
	Event      string                 `json:"event"`
	Timestamp  time.Time              `json:"timestamp"`
	DistinctID string                 `json:"distinct_id"`
	Properties map[string]interface{} `json:"properties"`
	PersonID   string                 `json:"person_id,omitempty"`
	UUID       string                 `json:"uuid,omitempty"`
}

// EventsResponse represents the API response for events list
type EventsResponse struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []Event `json:"results"`
}

// GetRecentEvents fetches recent events from the PostHog API
func (c *Client) GetRecentEvents(ctx context.Context, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 50
	}

	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	path := fmt.Sprintf("%s/events/?limit=%d&orderBy=-timestamp", c.getProjectPath(), limit)

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var eventsResp EventsResponse
	if err := json.Unmarshal(body, &eventsResp); err != nil {
		return nil, fmt.Errorf("failed to parse events response: %w", err)
	}

	return eventsResp.Results, nil
}

// GetEvent fetches a single event by ID
func (c *Client) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	path := fmt.Sprintf("%s/events/%s/", c.getProjectPath(), eventID)

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var event Event
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse event response: %w", err)
	}

	return &event, nil
}

// FormatEventTime formats an event timestamp for display
func FormatEventTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatEventTimeShort formats an event timestamp in short format
func FormatEventTimeShort(t time.Time) string {
	now := time.Now()
	if t.Format("2006-01-02") == now.Format("2006-01-02") {
		return t.Format("15:04:05")
	}
	return t.Format("01-02 15:04")
}
