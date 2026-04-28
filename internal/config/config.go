package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const configPath = "~/.config/lazyjira/config.yaml"

type Config struct {
	Server string `yaml:"server"`
	Email  string `yaml:"email"`
	Token  string `yaml:"token"`
}

func Load() (*Config, error) {
	path, err := expandPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolve config path %q: %w", configPath, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found at %s", path)
		}

		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config file %s: %w", path, err)
	}

	return &cfg, nil
}

func expandPath(path string) (string, error) {
	if path == "~" {
		return os.UserHomeDir()
	}

	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}

	return filepath.Join(home, path[2:]), nil
}

func (c *Config) validate() error {
	var missing []string

	if strings.TrimSpace(c.Server) == "" {
		missing = append(missing, "server")
	}
	if strings.TrimSpace(c.Email) == "" {
		missing = append(missing, "email")
	}
	if strings.TrimSpace(c.Token) == "" {
		missing = append(missing, "token")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required field(s): %s", strings.Join(missing, ", "))
	}

	return nil
}
