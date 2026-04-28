package jira

import (
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
