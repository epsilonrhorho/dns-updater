package ipify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ClientInterface defines the behavior for retrieving the public IP address.
type ClientInterface interface {
	GetIP(ctx context.Context) (string, error)
}

// Client implements the ClientInterface using an HTTP client.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient returns a new Client. If httpClient is nil, http.DefaultClient is used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    "https://api.ipify.org?format=json",
	}
}

// ipifyResponse represents the JSON structure returned by ipify.io.
type ipifyResponse struct {
	IP string `json:"ip"`
}

// GetIP fetches the public IP address.
func (c *Client) GetIP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var body ipifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}

	return body.IP, nil
}
