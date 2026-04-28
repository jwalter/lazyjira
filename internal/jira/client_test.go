package jira

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jwalter/lazyjira/internal/auth"
	"github.com/jwalter/lazyjira/internal/config"
)

func TestNewRequestSetsAuthorizationHeader(t *testing.T) {
	t.Parallel()

	client := New(config.Config{
		Server: "https://example.atlassian.net/",
		Email:  "user@example.com",
		Token:  "secret-token",
	})

	req, err := client.newRequest(http.MethodGet, "/rest/api/2/myself")
	if err != nil {
		t.Fatalf("newRequest() error = %v", err)
	}

	want := auth.BasicAuthHeader("user@example.com", "secret-token")
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}
}

func TestNewRequestConstructsURLCorrectly(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/rest/api/2/myself" {
			t.Fatalf("URL = %q, want %q", r.URL.String(), "/rest/api/2/myself")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(config.Config{
		Server: server.URL + "/",
		Email:  "user@example.com",
		Token:  "secret-token",
	})

	req, err := client.newRequest(http.MethodGet, "/rest/api/2/myself")
	if err == nil {
		resp, doErr := client.httpClient.Do(req)
		if doErr != nil {
			t.Fatalf("Do() error = %v", doErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
	}
	if err != nil {
		t.Fatalf("newRequest() error = %v", err)
	}
}

func TestGetCurrentUserSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want %q", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/rest/api/2/myself" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/rest/api/2/myself")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(config.Config{
		Server: server.URL,
		Email:  "user@example.com",
		Token:  "secret-token",
	})

	if err := client.GetCurrentUser(); err != nil {
		t.Fatalf("GetCurrentUser() error = %v", err)
	}
}

func TestGetCurrentUserFailureStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	client := New(config.Config{
		Server: server.URL,
		Email:  "user@example.com",
		Token:  "secret-token",
	})

	err := client.GetCurrentUser()
	if err == nil {
		t.Fatal("GetCurrentUser() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "unexpected status 403 Forbidden") {
		t.Fatalf("GetCurrentUser() error = %q, want unexpected status", err.Error())
	}
}

func TestSearchParsesValidResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want %q", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/rest/api/2/search" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/rest/api/2/search")
		}
		if got := r.URL.Query().Get("jql"); got != `project = TEST ORDER BY created DESC` {
			t.Fatalf("jql = %q, want %q", got, `project = TEST ORDER BY created DESC`)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"issues": []map[string]any{
				{
					"key": "TEST-1",
					"fields": map[string]any{
						"summary": "First issue",
						"status":  map[string]any{"name": "In Progress"},
					},
				},
				{
					"key": "TEST-2",
					"fields": map[string]any{
						"summary": "Second issue",
						"status":  map[string]any{"name": "Done"},
					},
				},
			},
		}); err != nil {
			t.Fatalf("Encode() error = %v", err)
		}
	}))
	defer server.Close()

	client := New(config.Config{
		Server: server.URL,
		Email:  "user@example.com",
		Token:  "secret-token",
	})

	issues, err := client.Search(`project = TEST ORDER BY created DESC`)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("len(issues) = %d, want 2", len(issues))
	}
	if issues[0] != (Issue{Key: "TEST-1", Summary: "First issue", Status: "In Progress"}) {
		t.Fatalf("issues[0] = %#v", issues[0])
	}
	if issues[1] != (Issue{Key: "TEST-2", Summary: "Second issue", Status: "Done"}) {
		t.Fatalf("issues[1] = %#v", issues[1])
	}
}

func TestSearchErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client := New(config.Config{
		Server: server.URL,
		Email:  "user@example.com",
		Token:  "secret-token",
	})

	_, err := client.Search("project = TEST")
	if err == nil {
		t.Fatal("Search() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "unexpected status 400 Bad Request") {
		t.Fatalf("Search() error = %q, want unexpected status", err.Error())
	}
}
