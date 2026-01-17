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

// GetRecentEvents fetches recent events using the Query API with HogQL
func (c *Client) GetRecentEvents(ctx context.Context, limit int) ([]Event, error) {
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
		if len(row) < 6 {
			continue
		}

		event := Event{}

		// UUID (column 0)
		if uuid, ok := row[0].(string); ok {
			event.UUID = uuid
			event.ID = uuid // Use UUID as ID
		}

		// Event name (column 1)
		if eventName, ok := row[1].(string); ok {
			event.Event = eventName
		}

		// Timestamp (column 2)
		if ts, ok := row[2].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
				event.Timestamp = parsed
			}
		}

		// Distinct ID (column 3)
		if distinctID, ok := row[3].(string); ok {
			event.DistinctID = distinctID
		}

		// Properties (column 4)
		if props, ok := row[4].(map[string]interface{}); ok {
			event.Properties = props
		}

		// Person ID (column 5)
		if personID, ok := row[5].(string); ok {
			event.PersonID = personID
		}

		events = append(events, event)
	}

	return events, nil
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
