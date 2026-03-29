package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProjects_PollingConfig(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		errContains string
		check       func(t *testing.T, projects []ProjectConfig)
	}{
		{
			name: "default poll_interval_ms",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
`,
			check: func(t *testing.T, projects []ProjectConfig) {
				if projects[0].PollIntervalMS != DefaultPollIntervalMS {
					t.Errorf("PollIntervalMS = %d, want %d", projects[0].PollIntervalMS, DefaultPollIntervalMS)
				}
			},
		},
		{
			name: "custom poll_interval_ms",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    poll_interval_ms: 5000
`,
			check: func(t *testing.T, projects []ProjectConfig) {
				if projects[0].PollIntervalMS != 5000 {
					t.Errorf("PollIntervalMS = %d, want %d", projects[0].PollIntervalMS, 5000)
				}
			},
		},
		{
			name: "zero poll_interval_ms is invalid",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    poll_interval_ms: 0
`,
			wantErr:     true,
			errContains: "poll_interval_ms must be positive",
		},
		{
			name: "negative poll_interval_ms is invalid",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    poll_interval_ms: -1
`,
			wantErr:     true,
			errContains: "poll_interval_ms must be positive",
		},
		{
			name: "default max_concurrent_tasks",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
`,
			check: func(t *testing.T, projects []ProjectConfig) {
				if projects[0].MaxConcurrentTasks != DefaultMaxConcurrentTasks {
					t.Errorf("MaxConcurrentTasks = %d, want %d", projects[0].MaxConcurrentTasks, DefaultMaxConcurrentTasks)
				}
			},
		},
		{
			name: "explicit zero max_concurrent_tasks",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    max_concurrent_tasks: 0
`,
			check: func(t *testing.T, projects []ProjectConfig) {
				if projects[0].MaxConcurrentTasks != 0 {
					t.Errorf("MaxConcurrentTasks = %d, want 0", projects[0].MaxConcurrentTasks)
				}
			},
		},
		{
			name: "custom max_concurrent_tasks",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    max_concurrent_tasks: 5
`,
			check: func(t *testing.T, projects []ProjectConfig) {
				if projects[0].MaxConcurrentTasks != 5 {
					t.Errorf("MaxConcurrentTasks = %d, want 5", projects[0].MaxConcurrentTasks)
				}
			},
		},
		{
			name: "negative max_concurrent_tasks is invalid",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    max_concurrent_tasks: -1
`,
			wantErr:     true,
			errContains: "max_concurrent_tasks must be non-negative",
		},
		{
			name: "default shutdown_timeout_ms",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
`,
			check: func(t *testing.T, projects []ProjectConfig) {
				if projects[0].ShutdownTimeoutMS != DefaultShutdownTimeoutMS {
					t.Errorf("ShutdownTimeoutMS = %d, want %d", projects[0].ShutdownTimeoutMS, DefaultShutdownTimeoutMS)
				}
			},
		},
		{
			name: "custom shutdown_timeout_ms",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    shutdown_timeout_ms: 60000
`,
			check: func(t *testing.T, projects []ProjectConfig) {
				if projects[0].ShutdownTimeoutMS != 60000 {
					t.Errorf("ShutdownTimeoutMS = %d, want %d", projects[0].ShutdownTimeoutMS, 60000)
				}
			},
		},
		{
			name: "zero shutdown_timeout_ms is invalid",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    shutdown_timeout_ms: 0
`,
			wantErr:     true,
			errContains: "shutdown_timeout_ms must be positive",
		},
		{
			name: "multiple projects with different poll intervals",
			yaml: `projects:
  - clickup_list_id: "list-1"
    github_owner: "owner"
    github_repo: "repo-a"
    poll_interval_ms: 3000
  - clickup_list_id: "list-2"
    github_owner: "owner"
    github_repo: "repo-b"
    poll_interval_ms: 7000
`,
			check: func(t *testing.T, projects []ProjectConfig) {
				if len(projects) != 2 {
					t.Fatalf("len = %d, want 2", len(projects))
				}
				if projects[0].PollIntervalMS != 3000 {
					t.Errorf("project[0].PollIntervalMS = %d, want 3000", projects[0].PollIntervalMS)
				}
				if projects[1].PollIntervalMS != 7000 {
					t.Errorf("project[1].PollIntervalMS = %d, want 7000", projects[1].PollIntervalMS)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "projects.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.yaml), 0o600); err != nil {
				t.Fatal(err)
			}

			projects, err := loadProjects(tmpFile)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, projects)
			}
		})
	}
}

func TestLoadProjects_SpecOutput(t *testing.T) {
	tests := []struct {
		name           string
		yaml           string
		wantSpecOutput string
		wantErr        bool
		errContains    string
	}{
		{
			name: "spec_output omitted defaults to clickup",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
`,
			wantSpecOutput: "clickup",
		},
		{
			name: "spec_output clickup explicit",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    spec_output: "clickup"
`,
			wantSpecOutput: "clickup",
		},
		{
			name: "spec_output repo",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    spec_output: "repo"
`,
			wantSpecOutput: "repo",
		},
		{
			name: "spec_output invalid value",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    spec_output: "invalid"
`,
			wantErr:     true,
			errContains: "invalid spec_output",
		},
		{
			name: "spec_output uppercase Repo is normalized",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    spec_output: "Repo"
`,
			wantSpecOutput: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := t.TempDir() + "/projects.yaml"
			if err := os.WriteFile(tmpFile, []byte(tt.yaml), 0o600); err != nil {
				t.Fatal(err)
			}

			projects, err := loadProjects(tmpFile)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(projects) != 1 {
				t.Fatalf("expected 1 project, got %d", len(projects))
			}
			if projects[0].SpecOutput != tt.wantSpecOutput {
				t.Errorf("SpecOutput = %q, want %q", projects[0].SpecOutput, tt.wantSpecOutput)
			}
		})
	}
}
