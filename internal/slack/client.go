// Package slack provides Slack API client and wrappers.
package slack

import (
	"context"
	"encoding/json"
	"fmt"

	slackapi "github.com/rneatherway/slack"
)

// Client wraps the rneatherway/slack client for Slack API access.
type Client struct {
	api    *slackapi.Client
	domain string
}

// NewClient creates a new Slack client for the given team domain.
// It automatically sets up cookie-based authentication from local Slack Desktop data.
func NewClient(domain string) (*Client, error) {
	c := slackapi.NewClient(domain)
	if err := c.WithCookieAuth(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	return &Client{api: c, domain: domain}, nil
}

// NewClientNoCreds creates a client without triggering credential import.
// Used for auth creds to handle the import step explicitly.
func NewClientNoCreds(domain string) *Client {
	return &Client{api: slackapi.NewClient(domain), domain: domain}
}

// ImportCreds triggers the cookie-based authentication flow (extracts from Slack Desktop).
func (c *Client) ImportCreds() error {
	return c.api.WithCookieAuth()
}

// API makes a POST request to the given Slack API method and unmarshals the response.
func (c *Client) API(ctx context.Context, method string, params map[string]string) (map[string]any, error) {
	body, err := c.api.API(ctx, "POST", method, params, nil)
	if err != nil {
		return nil, fmt.Errorf("slack API %s: %w", method, err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal %s response: %w", method, err)
	}

	if ok, _ := result["ok"].(bool); !ok {
		errMsg, _ := result["error"].(string)
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("slack API %s: %s", method, errMsg)
	}

	return result, nil
}

// Domain returns the workspace domain.
func (c *Client) Domain() string {
	return c.domain
}
