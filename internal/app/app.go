package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jwalter/lazyjira/internal/jira"
)

type issueFetcher interface {
	Search(jql string) ([]jira.Issue, error)
}

type Model struct {
	client  issueFetcher
	loading bool
	err     error
	issues  []jira.Issue
}

type issuesLoadedMsg struct {
	issues []jira.Issue
	err    error
}

func New(client issueFetcher) Model {
	return Model{
		client:  client,
		loading: true,
	}
}

func (m Model) Init() tea.Cmd {
	return m.fetchIssuesCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case issuesLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.issues = msg.issues
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.loading {
		return "Loading issues..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	return ""
}

func (m Model) fetchIssuesCmd() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return issuesLoadedMsg{err: fmt.Errorf("jira client is not configured")}
		}

		issues, err := m.client.Search("")
		return issuesLoadedMsg{
			issues: issues,
			err:    err,
		}
	}
}
