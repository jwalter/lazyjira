package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func (c *Client) Search(jql string) ([]Issue, error) {
	req, err := c.newRequest(http.MethodGet, "/rest/api/2/search?jql="+url.QueryEscape(jql))
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search issues: unexpected status %s", resp.Status)
	}

	var payload struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary string `json:"summary"`
				Status  struct {
					Name string `json:"name"`
				} `json:"status"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("search issues: decode response: %w", err)
	}

	issues := make([]Issue, 0, len(payload.Issues))
	for _, issue := range payload.Issues {
		issues = append(issues, Issue{
			Key:     issue.Key,
			Summary: issue.Fields.Summary,
			Status:  issue.Fields.Status.Name,
		})
	}

	return issues, nil
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
