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

// ListRecentEvents fetches recent events using the Query API with HogQL
func (c *Client) ListRecentEvents(ctx context.Context, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 50
	}

	// Use HogQL query to get recent events
	query := fmt.Sprintf(`
		SELECT
			uuid,
			event,
			timestamp,
			distinct_id,
			properties,
			person_id
		FROM events
		ORDER BY timestamp DESC
		LIMIT %d
	`, limit)

	result, err := c.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	// Map query results to Event structs
	events := make([]Event, 0, len(result.Results))
	for _, row := range result.Results {
		if event, ok := parseEventFromRow(row); ok {
			events = append(events, event)
		}
	}

	return events, nil
}

// GetEvent fetches a single event by ID
func (c *Client) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	if err := c.ensureProjectInitialized(ctx); err != nil {
		return nil, fmt.Errorf("GetEvent: %w", err)
	}

	path := fmt.Sprintf("%s/events/%s/", c.getProjectPath(), eventID)

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("GetEvent: %w", err)
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
