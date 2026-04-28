package jira

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jwalter/lazyjira/internal/config"
)

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/myself" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/rest/api/2/myself")
		}
		if got := r.Header.Get("Authorization"); got != basicAuth("user@example.com", "secret-token") {
			t.Fatalf("Authorization = %q, want %q", got, basicAuth("user@example.com", "secret-token"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(config.Config{
		Server: server.URL + "/",
		Email:  "user@example.com",
		Token:  "secret-token",
	})

	if err := client.GetCurrentUser(); err != nil {
		t.Fatalf("GetCurrentUser() error = %v", err)
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/search" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/rest/api/2/search")
		}
		if got := r.URL.Query().Get("jql"); got != `project = TEST ORDER BY created DESC` {
			t.Fatalf("jql = %q, want %q", got, `project = TEST ORDER BY created DESC`)
		}
		if got := r.Header.Get("Authorization"); got != basicAuth("user@example.com", "secret-token") {
			t.Fatalf("Authorization = %q, want %q", got, basicAuth("user@example.com", "secret-token"))
		}

		response := map[string]any{
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
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
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

func TestGetCurrentUserUnexpectedStatus(t *testing.T) {
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

func basicAuth(email, token string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(email+":"+token))
}
