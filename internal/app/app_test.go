package app

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jwalter/lazyjira/internal/jira"
)

type mockClient struct {
	issues []jira.Issue
	err    error
	jql    string
}

func (m *mockClient) Search(jql string) ([]jira.Issue, error) {
	m.jql = jql
	return m.issues, m.err
}

func TestNewInitialState(t *testing.T) {
	t.Parallel()

	model := New(&mockClient{})

	if !model.loading {
		t.Fatal("loading = false, want true")
	}
	if got := model.View(); got != "Loading issues..." {
		t.Fatalf("View() = %q, want %q", got, "Loading issues...")
	}
}

func TestInitTriggersFetchIssuesCmd(t *testing.T) {
	t.Parallel()

	client := &mockClient{
		issues: []jira.Issue{{Key: "TEST-1", Summary: "Issue", Status: "Open"}},
	}
	model := New(client)

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init() cmd = nil, want non-nil")
	}

	msg := cmd()
	loaded, ok := msg.(issuesLoadedMsg)
	if !ok {
		t.Fatalf("cmd() msg type = %T, want issuesLoadedMsg", msg)
	}
	if loaded.err != nil {
		t.Fatalf("loaded.err = %v, want nil", loaded.err)
	}
	if client.jql != "" {
		t.Fatalf("Search() jql = %q, want empty string", client.jql)
	}
	if len(loaded.issues) != 1 {
		t.Fatalf("len(loaded.issues) = %d, want 1", len(loaded.issues))
	}
}

func TestUpdateHandlesLoadedIssues(t *testing.T) {
	t.Parallel()

	model := New(&mockClient{})
	next, cmd := model.Update(issuesLoadedMsg{
		issues: []jira.Issue{{Key: "TEST-1", Summary: "Issue", Status: "Done"}},
	})
	if cmd != nil {
		t.Fatal("Update() cmd != nil, want nil")
	}

	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("Update() model type = %T, want Model", next)
	}
	if updated.loading {
		t.Fatal("loading = true, want false")
	}
	if updated.err != nil {
		t.Fatalf("err = %v, want nil", updated.err)
	}
	if len(updated.issues) != 1 {
		t.Fatalf("len(issues) = %d, want 1", len(updated.issues))
	}
}

func TestUpdateHandlesErrorState(t *testing.T) {
	t.Parallel()

	model := New(&mockClient{})
	next, _ := model.Update(issuesLoadedMsg{err: errors.New("boom")})

	updated := next.(Model)
	if updated.loading {
		t.Fatal("loading = true, want false")
	}
	if updated.err == nil {
		t.Fatal("err = nil, want error")
	}
	if got := updated.View(); got != "Error: boom" {
		t.Fatalf("View() = %q, want %q", got, "Error: boom")
	}
}

func TestUpdateHandlesQuitKeys(t *testing.T) {
	t.Parallel()

	model := New(&mockClient{})
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("Update() cmd = nil, want tea.Quit")
	}
}
