package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ClickUpAPIToken    string
	ClickUpListID      string
	GitHubPAT          string
	GitHubOwner        string
	GitHubRepo         string
	GitHubWorkflowFile string // default: "agent.yml"
	PollIntervalMS     int    // default: 10000
}

func Load() (*Config, error) {
	cfg := &Config{
		GitHubWorkflowFile: "agent.yml",
		PollIntervalMS:     10000,
	}

	required := map[string]*string{
		"CLICKUP_API_TOKEN": &cfg.ClickUpAPIToken,
		"CLICKUP_LIST_ID":   &cfg.ClickUpListID,
		"GITHUB_PAT":        &cfg.GitHubPAT,
		"GITHUB_OWNER":      &cfg.GitHubOwner,
		"GITHUB_REPO":       &cfg.GitHubRepo,
	}

	var missing []string
	for envKey, field := range required {
		v := os.Getenv(envKey)
		if v == "" {
			missing = append(missing, envKey)
		} else {
			*field = v
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if v := os.Getenv("GITHUB_WORKFLOW_FILE"); v != "" {
		cfg.GitHubWorkflowFile = v
	}

	if v := os.Getenv("POLL_INTERVAL_MS"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid POLL_INTERVAL_MS value %q: %w", v, err)
		}
		if parsed <= 0 {
			return nil, fmt.Errorf("POLL_INTERVAL_MS must be positive, got %d", parsed)
		}
		cfg.PollIntervalMS = parsed
	}

	return cfg, nil
}
