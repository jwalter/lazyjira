package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	writeConfigFile(t, `
server: https://example.atlassian.net
email: user@example.com
token: secret-token
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server != "https://example.atlassian.net" {
		t.Fatalf("Config.Server = %q, want %q", cfg.Server, "https://example.atlassian.net")
	}
	if cfg.Email != "user@example.com" {
		t.Fatalf("Config.Email = %q, want %q", cfg.Email, "user@example.com")
	}
	if cfg.Token != "secret-token" {
		t.Fatalf("Config.Token = %q, want %q", cfg.Token, "secret-token")
	}
}

func TestLoadMissingFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	writeConfigFile(t, `
server: https://example.atlassian.net
email: ""
`)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}

	want := "missing required field(s): email, token"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("Load() error = %q, want substring %q", err.Error(), want)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}

	configPath := filepath.Join(os.Getenv("HOME"), ".config", "lazyjira", "config.yaml")
	want := "config file not found at " + configPath
	if err.Error() != want {
		t.Fatalf("Load() error = %q, want %q", err.Error(), want)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("LAZYJIRA_SERVER", "https://override.atlassian.net")
	t.Setenv("LAZYJIRA_EMAIL", "override@example.com")
	t.Setenv("LAZYJIRA_TOKEN", "override-token")

	writeConfigFile(t, `
server: https://example.atlassian.net
email: user@example.com
token: secret-token
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server != "https://override.atlassian.net" {
		t.Fatalf("Config.Server = %q, want %q", cfg.Server, "https://override.atlassian.net")
	}
	if cfg.Email != "override@example.com" {
		t.Fatalf("Config.Email = %q, want %q", cfg.Email, "override@example.com")
	}
	if cfg.Token != "override-token" {
		t.Fatalf("Config.Token = %q, want %q", cfg.Token, "override-token")
	}
}

func TestLoadPartialEnvOverride(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("LAZYJIRA_TOKEN", "override-token")

	writeConfigFile(t, `
server: https://example.atlassian.net
email: user@example.com
token: secret-token
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server != "https://example.atlassian.net" {
		t.Fatalf("Config.Server = %q, want %q", cfg.Server, "https://example.atlassian.net")
	}
	if cfg.Email != "user@example.com" {
		t.Fatalf("Config.Email = %q, want %q", cfg.Email, "user@example.com")
	}
	if cfg.Token != "override-token" {
		t.Fatalf("Config.Token = %q, want %q", cfg.Token, "override-token")
	}
}

func writeConfigFile(t *testing.T, contents string) {
	t.Helper()

	path := filepath.Join(os.Getenv("HOME"), ".config", "lazyjira", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
