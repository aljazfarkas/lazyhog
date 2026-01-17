package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// QueryResult represents a HogQL query result
type QueryResult struct {
	Columns []string        `json:"columns"`
	Results [][]interface{} `json:"results"`
	Types   []string        `json:"types,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// HogQLQuery represents the inner HogQL query structure
type HogQLQuery struct {
	Kind  string `json:"kind"`
	Query string `json:"query"`
}

// QueryRequest represents the new Query API request format
type QueryRequest struct {
	Query HogQLQuery `json:"query"`
	Name  string     `json:"name,omitempty"`
}

// QueryResponse represents the API response for Query API
type QueryResponse struct {
	Results [][]interface{} `json:"results"`
	Columns []string        `json:"columns"`
	Types   [][]string      `json:"types"`
	Error   string          `json:"error"`
}

// ExecuteQuery executes a HogQL query using the new Query API
func (c *Client) ExecuteQuery(ctx context.Context, query string) (*QueryResult, error) {
	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	path := fmt.Sprintf("%s/query/", c.getProjectPath())

	reqData := QueryRequest{
		Query: HogQLQuery{
			Kind:  "HogQLQuery",
			Query: query,
		},
	}

	resp, err := c.post(ctx, path, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var queryResp QueryResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to parse query response: %w", err)
	}

	if queryResp.Error != "" {
		return nil, fmt.Errorf("query error: %s", queryResp.Error)
	}

	// Extract just the type strings from the tuples
	types := make([]string, len(queryResp.Types))
	for i, typeTuple := range queryResp.Types {
		if len(typeTuple) >= 2 {
			types[i] = typeTuple[1]
		}
	}

	result := &QueryResult{
		Columns: queryResp.Columns,
		Results: queryResp.Results,
		Types:   types,
	}

	return result, nil
}
