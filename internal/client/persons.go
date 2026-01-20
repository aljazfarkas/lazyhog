package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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
	if err := c.ensureProjectInitialized(ctx); err != nil {
		return nil, fmt.Errorf("GetPerson: %w", err)
	}

	// We need to search for the person by distinct_id
	path := fmt.Sprintf("%s/persons/?distinct_id=%s", c.getProjectPath(), url.QueryEscape(distinctID))

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("GetPerson: %w", err)
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

	// Map query results to Event structs
	events := make([]Event, 0, len(result.Results))
	for _, row := range result.Results {
		if event, ok := parseEventFromRow(row); ok {
			events = append(events, event)
		}
	}

	return events, nil
}

// ListPersons fetches a list of persons
func (c *Client) ListPersons(ctx context.Context, limit int) ([]Person, error) {
	if limit <= 0 {
		limit = 50
	}

	if err := c.ensureProjectInitialized(ctx); err != nil {
		return nil, fmt.Errorf("ListPersons: %w", err)
	}

	path := fmt.Sprintf("%s/persons/?limit=%d", c.getProjectPath(), limit)

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("ListPersons: %w", err)
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
