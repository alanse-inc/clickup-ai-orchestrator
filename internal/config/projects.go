package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const defaultWorkflowFile = "agent.yml"

// ProjectConfig は1つの ClickUp リスト - GitHub リポジトリペアの設定
type ProjectConfig struct {
	ClickUpListID      string `yaml:"clickup_list_id"`
	GitHubOwner        string `yaml:"github_owner"`
	GitHubRepo         string `yaml:"github_repo"`
	GitHubWorkflowFile string `yaml:"github_workflow_file"`
}

type projectsFile struct {
	Projects []ProjectConfig `yaml:"projects"`
}

func loadProjects(path string) ([]ProjectConfig, error) {
	data, err := os.ReadFile(path) //nolint:gosec // パスは環境変数 PROJECTS_FILE またはデフォルト値で制御される
	if err != nil {
		return nil, fmt.Errorf("reading projects file: %w", err)
	}

	var pf projectsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("parsing projects file: %w", err)
	}

	if len(pf.Projects) == 0 {
		return nil, fmt.Errorf("projects file must contain at least one project")
	}

	for i, p := range pf.Projects {
		var missing []string
		if p.ClickUpListID == "" {
			missing = append(missing, "clickup_list_id")
		}
		if p.GitHubOwner == "" {
			missing = append(missing, "github_owner")
		}
		if p.GitHubRepo == "" {
			missing = append(missing, "github_repo")
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("project[%d]: missing required fields: %v", i, missing)
		}
		if p.GitHubWorkflowFile == "" {
			pf.Projects[i].GitHubWorkflowFile = defaultWorkflowFile
		}
	}

	return pf.Projects, nil
}
