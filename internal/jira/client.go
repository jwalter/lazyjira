package jira

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jwalter/lazyjira/internal/auth"
	"github.com/jwalter/lazyjira/internal/config"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	authHeader string
}

type Issue struct {
	Key     string
	Summary string
	Status  string
}

func New(cfg config.Config) *Client {
	return &Client{
		baseURL:    strings.TrimRight(cfg.Server, "/"),
		httpClient: http.DefaultClient,
		authHeader: auth.BasicAuthHeader(cfg.Email, cfg.Token),
	}
}

func (c *Client) GetCurrentUser() error {
	req, err := c.newRequest(http.MethodGet, "/rest/api/2/myself")
	if err != nil {
		return fmt.Errorf("get current user: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("get current user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get current user: unexpected status %s", resp.Status)
	}

	return nil
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	if strings.TrimSpace(c.baseURL) == "" {
		return nil, fmt.Errorf("build request: base URL is empty")
	}

	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
