package onec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for communicating with 1C:Enterprise.
type Client struct {
	BaseURL    string
	User       string
	Password   string
	HTTPClient *http.Client
}

// NewClient creates a client for 1C HTTP service.
// When user is non-empty, basic auth is added to every request.
func NewClient(baseURL, user, password string) *Client {
	return &Client{
		BaseURL:  baseURL,
		User:     user,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get performs a GET request to a 1C endpoint and decodes the JSON response.
func (c *Client) Get(ctx context.Context, endpoint string, result any) error {
	url := c.BaseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if c.User != "" {
		req.SetBasicAuth(c.User, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request to 1C: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("1C returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding 1C response: %w", err)
	}

	return nil
}

// Post performs a POST request to a 1C endpoint with a JSON body and decodes the JSON response.
func (c *Client) Post(ctx context.Context, endpoint string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	url := c.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if c.User != "" {
		req.SetBasicAuth(c.User, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request to 1C: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("1C returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding 1C response: %w", err)
	}

	return nil
}
