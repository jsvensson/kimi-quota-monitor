package kimi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client calls the Kimi Code API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a Client for the given base URL and API key.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Usages fetches the current quota usage.
func (c *Client) Usages(ctx context.Context) (Usages, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/usages", nil)
	if err != nil {
		return Usages{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Usages{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Usages{}, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var usages Usages
	if err := json.NewDecoder(resp.Body).Decode(&usages); err != nil {
		return Usages{}, fmt.Errorf("decode response: %w", err)
	}
	return usages, nil
}
