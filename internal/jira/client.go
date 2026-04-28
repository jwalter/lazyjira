package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/jwalter/lazyjira/internal/config"
)

const (
	myselfPath = "/rest/api/2/myself"
	searchPath = "/rest/api/2/search"
)

type Client struct {
	baseURL    *url.URL
	email      string
	token      string
	httpClient *http.Client
}

type Issue struct {
	Key     string
	Summary string
	Status  string
}

func New(cfg config.Config) *Client {
	baseURL, _ := url.Parse(strings.TrimRight(cfg.Server, "/"))

	return &Client{
		baseURL: baseURL,
		email:   cfg.Email,
		token:   cfg.Token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetCurrentUser() error {
	req, err := c.newRequest(http.MethodGet, myselfPath, nil)
	if err != nil {
		return err
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
	query := url.Values{}
	query.Set("jql", jql)

	req, err := c.newRequest(http.MethodGet, searchPath, query)
	if err != nil {
		return nil, err
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

func (c *Client) newRequest(method, requestPath string, query url.Values) (*http.Request, error) {
	if c.baseURL == nil {
		return nil, fmt.Errorf("build request: invalid base URL")
	}

	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, requestPath)
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.SetBasicAuth(c.email, c.token)
	req.Header.Set("Accept", "application/json")

	return req, nil
}
