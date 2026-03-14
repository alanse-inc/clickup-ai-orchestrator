package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Dispatcher struct {
	pat          string
	owner        string
	repo         string
	workflowFile string
	httpClient   *http.Client
}

func NewDispatcher(pat, owner, repo, workflowFile string) *Dispatcher {
	return &Dispatcher{
		pat:          pat,
		owner:        owner,
		repo:         repo,
		workflowFile: workflowFile,
		httpClient:   &http.Client{},
	}
}

type dispatchRequest struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs"`
}

func (d *Dispatcher) TriggerWorkflow(ctx context.Context, taskID string, phase string, statusOnSuccess string, statusOnError string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches",
		d.owner, d.repo, d.workflowFile)

	body := dispatchRequest{
		Ref: "main",
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

	req.Header.Set("Authorization", "Bearer "+d.pat)
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
