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
	// We need to search for the person by distinct_id
	path := fmt.Sprintf("/api/projects/@current/persons/?distinct_id=%s", url.QueryEscape(distinctID))

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

// GetPersonEvents fetches recent events for a person
func (c *Client) GetPersonEvents(ctx context.Context, distinctID string, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 10
	}

	path := fmt.Sprintf("/api/projects/@current/events/?person_id=%s&limit=%d&orderBy=-timestamp",
		url.QueryEscape(distinctID), limit)

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
