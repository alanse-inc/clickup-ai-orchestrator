package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultWorkflowRef = "main"
	defaultHTTPTimeout = 30 * time.Second
)

// Dispatcher は GitHub Actions の workflow_dispatch イベントをトリガーする
type Dispatcher struct {
	auth         Authenticator
	owner        string
	repo         string
	workflowFile string
	httpClient   *http.Client
}

// NewDispatcher は新しい Dispatcher を生成する
func NewDispatcher(auth Authenticator, owner, repo, workflowFile string) *Dispatcher {
	return &Dispatcher{
		auth:         auth,
		owner:        owner,
		repo:         repo,
		workflowFile: workflowFile,
		httpClient:   &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// Ping は GitHub API への疎通を確認する。
// GET /rate_limit を呼び出し、認証が有効かつ API が到達可能かを検証する。
func (d *Dispatcher) Ping(ctx context.Context) error {
	url := githubAPIBase + "/rate_limit"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating ping request: %w", err)
	}
	if err := d.auth.SetAuth(req); err != nil {
		return fmt.Errorf("setting auth: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pinging GitHub API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

type dispatchRequest struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs"`
}

// TriggerWorkflow は指定したタスクのフェーズに対応する GitHub Actions ワークフローをトリガーする。
// statusOnSuccess と statusOnError はワークフロー完了後に設定されるステータスとして inputs に渡される。
func (d *Dispatcher) TriggerWorkflow(ctx context.Context, taskID string, phase string, statusOnSuccess string, statusOnError string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/actions/workflows/%s/dispatches",
		githubAPIBase, d.owner, d.repo, d.workflowFile)

	body := dispatchRequest{
		Ref: defaultWorkflowRef,
		Inputs: map[string]string{
			"task_id":           taskID,
			"phase":             phase,
			"status_on_success": statusOnSuccess,
			"status_on_error":   statusOnError,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := d.auth.SetAuth(req); err != nil {
		return fmt.Errorf("failed to authenticate request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
