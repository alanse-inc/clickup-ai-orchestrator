package config

import (
	"os"
	"strings"
	"testing"
)

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
			name: "spec_output uppercase Repo is invalid",
			yaml: `projects:
  - clickup_list_id: "list-123"
    github_owner: "owner"
    github_repo: "repo"
    spec_output: "Repo"
`,
			wantErr:     true,
			errContains: "invalid spec_output",
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
