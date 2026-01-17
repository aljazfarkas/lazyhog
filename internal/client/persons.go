package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"
)

// Person represents a PostHog person
type Person struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	DistinctIDs []string               `json:"distinct_ids"`
	Properties  map[string]interface{} `json:"properties"`
	CreatedAt   string                 `json:"created_at"`
	UUID        string                 `json:"uuid"`
}

// PersonsResponse represents the API response for persons list
type PersonsResponse struct {
	Next     *string  `json:"next"`
	Previous *string  `json:"previous"`
	Results  []Person `json:"results"`
}

// GetPerson fetches a person by distinct ID
func (c *Client) GetPerson(ctx context.Context, distinctID string) (*Person, error) {
	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	// We need to search for the person by distinct_id
	path := fmt.Sprintf("%s/persons/?distinct_id=%s", c.getProjectPath(), url.QueryEscape(distinctID))

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var personsResp PersonsResponse
	if err := json.Unmarshal(body, &personsResp); err != nil {
		return nil, fmt.Errorf("failed to parse persons response: %w", err)
	}

	if len(personsResp.Results) == 0 {
		return nil, fmt.Errorf("person not found with distinct_id: %s", distinctID)
	}

	return &personsResp.Results[0], nil
}

// GetPersonEvents fetches recent events for a person using Query API with HogQL
func (c *Client) GetPersonEvents(ctx context.Context, distinctID string, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 10
	}

	// Use HogQL query to get events for a specific person
	query := fmt.Sprintf(`
		SELECT
			uuid,
			event,
			timestamp,
			distinct_id,
			properties,
			person_id
		FROM events
		WHERE distinct_id = '%s'
		ORDER BY timestamp DESC
		LIMIT %d
	`, distinctID, limit)

	result, err := c.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query person events: %w", err)
	}

	// Map query results to Event structs using the same logic as GetRecentEvents
	events := make([]Event, 0, len(result.Results))
	for _, row := range result.Results {
		if len(row) < 6 {
			continue
		}

		event := Event{}

		// UUID (column 0)
		if uuid, ok := row[0].(string); ok {
			event.UUID = uuid
			event.ID = uuid
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

// ListPersons fetches a list of persons
func (c *Client) ListPersons(ctx context.Context, limit int) ([]Person, error) {
	if limit <= 0 {
		limit = 50
	}

	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	path := fmt.Sprintf("%s/persons/?limit=%d", c.getProjectPath(), limit)

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var personsResp PersonsResponse
	if err := json.Unmarshal(body, &personsResp); err != nil {
		return nil, fmt.Errorf("failed to parse persons response: %w", err)
	}

	return personsResp.Results, nil
}
