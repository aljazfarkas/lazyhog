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

// HogQLQueryRequest represents a HogQL query request
type HogQLQueryRequest struct {
	Query string `json:"query"`
}

// HogQLQueryResponse represents the API response for HogQL queries
type HogQLQueryResponse struct {
	Results [][]interface{} `json:"results"`
	Columns []string        `json:"columns"`
	Types   []string        `json:"types"`
	Error   string          `json:"error"`
}

// ExecuteQuery executes a HogQL query
func (c *Client) ExecuteQuery(ctx context.Context, query string) (*QueryResult, error) {
	// Ensure project ID is initialized
	if c.projectID == 0 {
		if err := c.InitializeProject(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize project: %w", err)
		}
	}

	path := fmt.Sprintf("%s/query/", c.getProjectPath())

	reqData := HogQLQueryRequest{
		Query: query,
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

	var queryResp HogQLQueryResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to parse query response: %w", err)
	}

	if queryResp.Error != "" {
		return nil, fmt.Errorf("query error: %s", queryResp.Error)
	}

	result := &QueryResult{
		Columns: queryResp.Columns,
		Results: queryResp.Results,
		Types:   queryResp.Types,
	}

	return result, nil
}
